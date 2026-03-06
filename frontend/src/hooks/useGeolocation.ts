import { useState, useCallback } from "react";

interface LocationState {
  latitude: number;
  longitude: number;
  name: string;
}

export function useGeolocation(initialLocation?: LocationState | null) {
  const [location, setLocation] = useState<LocationState | null>(initialLocation ?? null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const requestLocation = useCallback(async () => {
    if (!navigator.geolocation) {
      setError("이 브라우저에서는 위치 서비스를 사용할 수 없습니다.");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const position = await new Promise<GeolocationPosition>((resolve, reject) => {
        navigator.geolocation.getCurrentPosition(resolve, reject, {
          enableHighAccuracy: false,
          timeout: 10000,
          maximumAge: 300000,
        });
      });

      const lat = Math.round(position.coords.latitude * 100) / 100;
      const lng = Math.round(position.coords.longitude * 100) / 100;

      let name = "알 수 없는 위치";
      try {
        const res = await fetch(
          `https://nominatim.openstreetmap.org/reverse?format=json&lat=${lat}&lon=${lng}&zoom=14&accept-language=ko`,
        );
        if (res.ok) {
          const data = await res.json();
          const addr = data.address;
          if (addr) {
            const parts = [
              addr.city || addr.town || addr.village || addr.county,
              addr.suburb || addr.neighbourhood || addr.quarter,
            ].filter(Boolean);
            name = parts.join(" ") || data.display_name?.split(",").slice(0, 2).join(", ") || name;
          }
        }
      } catch {
        // Reverse geocoding failed, use default name
      }

      setLocation({ latitude: lat, longitude: lng, name });
    } catch (err) {
      if (err instanceof GeolocationPositionError) {
        switch (err.code) {
          case err.PERMISSION_DENIED:
            setError("위치 접근이 거부되었습니다.");
            break;
          case err.POSITION_UNAVAILABLE:
            setError("위치 정보를 사용할 수 없습니다.");
            break;
          case err.TIMEOUT:
            setError("위치 요청 시간이 초과되었습니다.");
            break;
        }
      } else {
        setError("위치를 가져오는 중 오류가 발생했습니다.");
      }
    } finally {
      setIsLoading(false);
    }
  }, []);

  const updateName = useCallback((name: string) => {
    setLocation((prev) => (prev ? { ...prev, name: name.slice(0, 100) } : null));
  }, []);

  const clearLocation = useCallback(() => {
    setLocation(null);
    setError(null);
  }, []);

  return {
    location,
    isLoading,
    error,
    requestLocation,
    updateName,
    clearLocation,
  };
}

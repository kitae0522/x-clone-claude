const MINUTE = 60;
const HOUR = 3600;
const DAY = 86400;

export function formatRelativeTime(dateString: string): string {
  const now = Date.now();
  const then = new Date(dateString).getTime();
  const diffSec = Math.floor((now - then) / 1000);

  if (diffSec < MINUTE) return "방금";
  if (diffSec < HOUR) return `${Math.floor(diffSec / MINUTE)}분`;
  if (diffSec < DAY) return `${Math.floor(diffSec / HOUR)}시간`;
  if (diffSec < DAY * 7) return `${Math.floor(diffSec / DAY)}일`;

  const date = new Date(dateString);
  const thisYear = new Date().getFullYear();

  if (date.getFullYear() === thisYear) {
    return `${date.getMonth() + 1}월 ${date.getDate()}일`;
  }

  return `${date.getFullYear()}년 ${date.getMonth() + 1}월 ${date.getDate()}일`;
}

export function formatCompactNumber(num: number): string {
  if (num < 1000) return num.toString();
  if (num < 1_000_000) {
    const k = num / 1000;
    return k >= 100
      ? `${Math.floor(k)}K`
      : `${k.toFixed(1).replace(/\.0$/, "")}K`;
  }
  const m = num / 1_000_000;
  return m >= 100
    ? `${Math.floor(m)}M`
    : `${m.toFixed(1).replace(/\.0$/, "")}M`;
}

import { useState, useRef, useEffect } from "react";
import { Globe, Users, Lock, ChevronDown } from "lucide-react";

type Visibility = "public" | "follower" | "private";

interface VisibilitySelectorProps {
  value: Visibility;
  onChange: (value: Visibility) => void;
  disabled?: boolean;
}

const options = [
  {
    value: "public" as Visibility,
    icon: Globe,
    label: "전체 공개",
    description: "모든 사용자가 볼 수 있습니다",
  },
  {
    value: "follower" as Visibility,
    icon: Users,
    label: "팔로워 전용",
    description: "나를 팔로우하는 사용자만 볼 수 있습니다",
  },
  {
    value: "private" as Visibility,
    icon: Lock,
    label: "나만 보기",
    description: "나만 볼 수 있습니다",
  },
];

export default function VisibilitySelector({
  value,
  onChange,
  disabled,
}: VisibilitySelectorProps) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  const selected = options.find((o) => o.value === value) ?? options[0];
  const Icon = selected.icon;

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    if (open) {
      document.addEventListener("mousedown", handleClickOutside);
      return () =>
        document.removeEventListener("mousedown", handleClickOutside);
    }
  }, [open]);

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => !disabled && setOpen(!open)}
        disabled={disabled}
        className="flex cursor-pointer items-center gap-1.5 rounded-full border-none bg-transparent px-3 py-1.5 text-[13px] font-medium text-primary transition-colors hover:bg-primary/10 disabled:opacity-50"
      >
        <Icon size={14} />
        <span>{selected.label}</span>
        <ChevronDown size={12} />
      </button>

      {open && (
        <div className="absolute left-0 top-full z-50 mt-1 w-64 overflow-hidden rounded-xl border border-border bg-background shadow-lg">
          {options.map((option) => {
            const OptionIcon = option.icon;
            const isSelected = option.value === value;
            return (
              <button
                key={option.value}
                type="button"
                onClick={() => {
                  onChange(option.value);
                  setOpen(false);
                }}
                className={`flex w-full cursor-pointer items-center gap-3 border-none px-4 py-3 text-left transition-colors hover:bg-foreground/5 ${
                  isSelected ? "bg-primary/5" : "bg-transparent"
                }`}
              >
                <OptionIcon
                  size={18}
                  className={isSelected ? "text-primary" : "text-muted-foreground"}
                />
                <div>
                  <div
                    className={`text-[14px] ${isSelected ? "font-semibold text-primary" : "text-foreground"}`}
                  >
                    {option.label}
                  </div>
                  <div className="text-[12px] text-muted-foreground">
                    {option.description}
                  </div>
                </div>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}

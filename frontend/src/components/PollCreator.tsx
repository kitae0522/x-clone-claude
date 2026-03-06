import { Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";

const DURATION_OPTIONS = [
  { label: "1시간", value: 60 },
  { label: "6시간", value: 360 },
  { label: "12시간", value: 720 },
  { label: "1일", value: 1440 },
  { label: "3일", value: 4320 },
  { label: "7일", value: 10080 },
];

interface PollCreatorProps {
  options: string[];
  durationMinutes: number;
  onOptionsChange: (options: string[]) => void;
  onDurationChange: (minutes: number) => void;
  onRemove: () => void;
}

export default function PollCreator({
  options,
  durationMinutes,
  onOptionsChange,
  onDurationChange,
  onRemove,
}: PollCreatorProps) {
  function handleOptionChange(index: number, value: string) {
    const next = [...options];
    next[index] = value.slice(0, 25);
    onOptionsChange(next);
  }

  function addOption() {
    if (options.length < 4) {
      onOptionsChange([...options, ""]);
    }
  }

  function removeOption(index: number) {
    if (options.length > 2) {
      onOptionsChange(options.filter((_, i) => i !== index));
    }
  }

  return (
    <div className="mt-2 rounded-xl border border-border p-3">
      <div className="mb-2 flex items-center justify-between">
        <span className="text-sm font-medium text-foreground">투표 만들기</span>
        <button
          type="button"
          onClick={onRemove}
          className="cursor-pointer rounded-full border-none bg-transparent p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        >
          <X size={16} />
        </button>
      </div>

      <div className="space-y-2">
        {options.map((option, index) => (
          <div key={index} className="flex items-center gap-2">
            <input
              type="text"
              value={option}
              onChange={(e) => handleOptionChange(index, e.target.value)}
              placeholder={`선택지 ${index + 1}`}
              maxLength={25}
              className="flex-1 rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none"
            />
            <span className="shrink-0 text-[11px] text-muted-foreground">
              {option.length}/25
            </span>
            {options.length > 2 && (
              <button
                type="button"
                onClick={() => removeOption(index)}
                className="cursor-pointer rounded-full border-none bg-transparent p-1 text-muted-foreground transition-colors hover:text-destructive"
              >
                <X size={14} />
              </button>
            )}
          </div>
        ))}
      </div>

      {options.length < 4 && (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={addOption}
          className="mt-2 w-full text-primary"
        >
          <Plus size={16} className="mr-1" />
          선택지 추가
        </Button>
      )}

      <div className="mt-3 border-t border-border pt-3">
        <label className="mb-1 block text-xs text-muted-foreground">투표 기간</label>
        <select
          value={durationMinutes}
          onChange={(e) => onDurationChange(Number(e.target.value))}
          className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none"
        >
          {DURATION_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value} className="bg-background text-foreground">
              {opt.label}
            </option>
          ))}
        </select>
      </div>
    </div>
  );
}

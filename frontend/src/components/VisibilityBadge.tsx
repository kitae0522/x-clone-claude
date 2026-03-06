import { Users, Lock } from "lucide-react";

interface VisibilityBadgeProps {
  visibility: "public" | "follower" | "private";
}

export default function VisibilityBadge({ visibility }: VisibilityBadgeProps) {
  if (visibility === "public") return null;

  if (visibility === "follower") {
    return (
      <span className="inline-flex items-center gap-1 text-[13px] text-muted-foreground">
        <span>·</span>
        <Users size={12} />
        <span>팔로워 전용</span>
      </span>
    );
  }

  return (
    <span className="inline-flex items-center gap-1 text-[13px] text-muted-foreground">
      <span>·</span>
      <Lock size={12} />
      <span>나만 보기</span>
    </span>
  );
}

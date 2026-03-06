import { useState } from "react";
import type { PollData } from "@/types/api";
import { useAuth } from "@/hooks/useAuthContext";
import { useVote } from "@/hooks/usePoll";
import { cn } from "@/lib/utils";
import { Check } from "lucide-react";

interface PollDisplayProps {
  poll: PollData;
  postId: string;
}

export default function PollDisplay({ poll, postId }: PollDisplayProps) {
  const { user } = useAuth();
  const voteMutation = useVote(postId);
  const [optimisticVotedIndex, setOptimisticVotedIndex] = useState<
    number | null
  >(null);

  const hasVoted = poll.votedIndex >= 0 || optimisticVotedIndex !== null;
  const showResults = hasVoted || poll.isExpired;
  const votedIndex = optimisticVotedIndex ?? poll.votedIndex;

  const expiresAt = new Date(poll.expiresAt);
  const now = new Date();
  const timeLeft = expiresAt.getTime() - now.getTime();

  function formatTimeLeft(): string {
    if (poll.isExpired || timeLeft <= 0) return "투표 종료";
    const hours = Math.floor(timeLeft / (1000 * 60 * 60));
    const minutes = Math.floor((timeLeft % (1000 * 60 * 60)) / (1000 * 60));
    if (hours >= 24) {
      const days = Math.floor(hours / 24);
      return `${days}일 남음`;
    }
    if (hours > 0) return `${hours}시간 ${minutes}분 남음`;
    return `${minutes}분 남음`;
  }

  function handleVote(e: React.MouseEvent, optionIndex: number) {
    e.stopPropagation();
    if (!user || hasVoted || poll.isExpired) return;

    setOptimisticVotedIndex(optionIndex);
    voteMutation.mutate(optionIndex);
  }

  const totalVotes =
    optimisticVotedIndex !== null ? poll.totalVotes + 1 : poll.totalVotes;

  return (
    <div className="mt-3 space-y-2" onClick={(e) => e.stopPropagation()}>
      {poll.options.map((option, index) => {
        const voteCount =
          optimisticVotedIndex === index
            ? option.voteCount + 1
            : option.voteCount;
        const percentage =
          totalVotes > 0 ? Math.round((voteCount / totalVotes) * 100) : 0;
        const isSelected = votedIndex === index;
        const isWinning =
          showResults &&
          voteCount === Math.max(...poll.options.map((o) => o.voteCount));

        if (showResults) {
          return (
            <div
              key={index}
              className="relative overflow-hidden rounded-lg border border-border p-3"
            >
              <div
                className={cn(
                  "absolute inset-0 rounded-lg transition-all",
                  isWinning ? "bg-primary/15" : "bg-muted/50",
                )}
                style={{ width: `${percentage}%` }}
              />
              <div className="relative flex items-center justify-between">
                <div className="flex items-center gap-2">
                  {isSelected && <Check size={16} className="text-primary" />}
                  <span className={cn("text-sm", isWinning && "font-bold")}>
                    {option.text}
                  </span>
                </div>
                <span
                  className={cn(
                    "text-sm",
                    isWinning ? "font-bold" : "text-muted-foreground",
                  )}
                >
                  {percentage}%
                </span>
              </div>
            </div>
          );
        }

        return (
          <button
            key={index}
            onClick={(e) => handleVote(e, index)}
            disabled={!user || voteMutation.isPending}
            className="w-full cursor-pointer rounded-lg border border-primary/50 bg-transparent p-3 text-center text-sm font-medium text-primary transition-colors hover:bg-primary/10 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {option.text}
          </button>
        );
      })}

      <div className="flex items-center gap-2 text-[13px] text-muted-foreground">
        <span>{totalVotes}표</span>
        <span>·</span>
        <span>{formatTimeLeft()}</span>
      </div>
    </div>
  );
}

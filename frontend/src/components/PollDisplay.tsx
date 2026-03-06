import { useState } from "react";
import type { PollData } from "@/types/api";
import { useAuth } from "@/hooks/useAuthContext";
import { useVote, useUnvote } from "@/hooks/usePoll";
import { cn } from "@/lib/utils";
import { Check, RotateCcw } from "lucide-react";

interface PollDisplayProps {
  poll: PollData;
  postId: string;
}

export default function PollDisplay({ poll, postId }: PollDisplayProps) {
  const { user } = useAuth();
  const voteMutation = useVote(postId);
  const unvoteMutation = useUnvote(postId);
  const [optimisticVotedIndex, setOptimisticVotedIndex] = useState<
    number | null
  >(null);
  const [optimisticUnvoted, setOptimisticUnvoted] = useState(false);

  const hasVoted =
    !optimisticUnvoted &&
    (poll.votedIndex >= 0 || optimisticVotedIndex !== null);
  const showResults = hasVoted || poll.isExpired;
  const votedIndex = optimisticUnvoted
    ? -1
    : (optimisticVotedIndex ?? poll.votedIndex);

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
    setOptimisticUnvoted(false);
    voteMutation.mutate(optionIndex);
  }

  function handleUnvote(e: React.MouseEvent) {
    e.stopPropagation();
    if (!user || !hasVoted || poll.isExpired) return;

    setOptimisticUnvoted(true);
    setOptimisticVotedIndex(null);
    unvoteMutation.mutate();
  }

  const totalVotes = optimisticUnvoted
    ? Math.max(0, poll.totalVotes - 1)
    : optimisticVotedIndex !== null
      ? poll.totalVotes + 1
      : poll.totalVotes;

  const maxVoteCount = Math.max(
    ...poll.options.map((o, i) => {
      if (optimisticUnvoted && poll.votedIndex === i)
        return Math.max(0, o.voteCount - 1);
      if (optimisticVotedIndex === i) return o.voteCount + 1;
      return o.voteCount;
    }),
  );

  return (
    <div className="mt-3 space-y-2" onClick={(e) => e.stopPropagation()}>
      {poll.options.map((option, index) => {
        let voteCount = option.voteCount;
        if (optimisticUnvoted && poll.votedIndex === index) {
          voteCount = Math.max(0, voteCount - 1);
        } else if (optimisticVotedIndex === index) {
          voteCount = voteCount + 1;
        }

        const percentage =
          totalVotes > 0 ? Math.round((voteCount / totalVotes) * 100) : 0;
        const isSelected = votedIndex === index;
        const isWinning = showResults && voteCount === maxVoteCount && voteCount > 0;

        if (showResults) {
          return (
            <div
              key={index}
              className={cn(
                "relative overflow-hidden rounded-lg border p-3 transition-all",
                isSelected
                  ? "border-primary/50"
                  : "border-border",
              )}
            >
              {/* Progress bar background */}
              <div
                className={cn(
                  "absolute inset-y-0 left-0 rounded-lg transition-all duration-500 ease-out",
                  isWinning
                    ? "bg-primary/20"
                    : "bg-muted/60",
                )}
                style={{ width: `${percentage}%` }}
              />
              <div className="relative flex items-center justify-between">
                <div className="flex items-center gap-2">
                  {isSelected && (
                    <Check size={16} className="shrink-0 text-primary" />
                  )}
                  <span
                    className={cn(
                      "text-sm",
                      isWinning && "font-semibold",
                      isSelected && "text-primary",
                    )}
                  >
                    {option.text}
                  </span>
                </div>
                <span
                  className={cn(
                    "ml-2 shrink-0 text-sm tabular-nums",
                    isWinning
                      ? "font-semibold text-foreground"
                      : "text-muted-foreground",
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

      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 text-[13px] text-muted-foreground">
          <span>{totalVotes}표</span>
          <span>·</span>
          <span>{formatTimeLeft()}</span>
        </div>
        {hasVoted && !poll.isExpired && (
          <button
            onClick={handleUnvote}
            disabled={unvoteMutation.isPending}
            className="flex cursor-pointer items-center gap-1 rounded-full border-none bg-transparent px-2 py-1 text-[13px] text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:opacity-50"
          >
            <RotateCcw size={12} />
            다시 투표
          </button>
        )}
      </div>
    </div>
  );
}

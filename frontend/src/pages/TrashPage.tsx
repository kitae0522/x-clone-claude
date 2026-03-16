import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowLeft, RotateCcw, Trash2 } from "lucide-react";
import { useTrash, useRestorePost, usePermanentDelete } from "@/hooks/useTrash";
import { Button } from "@/components/ui/button";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import UserAvatar from "@/components/UserAvatar";
import { formatRelativeTime } from "@/lib/formatTime";
import type { TrashPost } from "@/types/api";

export default function TrashPage() {
  const navigate = useNavigate();
  const { data, isLoading, hasNextPage, fetchNextPage, isFetchingNextPage } =
    useTrash();
  const restoreMutation = useRestorePost();
  const permanentDeleteMutation = usePermanentDelete();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const posts = data?.pages.flatMap((page) => page.posts) ?? [];

  return (
    <>
      <header className="sticky top-0 z-10 flex items-center gap-4 border-b border-border bg-background/65 px-4 py-3 backdrop-blur-xl">
        <button
          className="cursor-pointer rounded-full border-none bg-transparent p-2 text-foreground transition-colors hover:bg-foreground/10"
          onClick={() => navigate(-1)}
        >
          <ArrowLeft className="h-5 w-5" />
        </button>
        <h1 className="text-xl font-bold">휴지통</h1>
      </header>

      <div className="px-4 py-3 text-sm text-muted-foreground border-b border-border">
        삭제된 게시글은 30일 후 자동으로 영구 삭제됩니다.
      </div>

      {isLoading ? (
        <div className="flex justify-center py-8">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      ) : posts.length === 0 ? (
        <p className="px-4 py-12 text-center text-muted-foreground">
          휴지통이 비어있습니다.
        </p>
      ) : (
        <div>
          {posts.map((post) => (
            <TrashItem
              key={post.id}
              post={post}
              onRestore={(id) => restoreMutation.mutate(id)}
              onPermanentDelete={(id) => setDeleteTarget(id)}
              isRestoring={restoreMutation.isPending}
            />
          ))}
          {hasNextPage && (
            <div className="flex justify-center py-4">
              <Button
                variant="outline"
                onClick={() => fetchNextPage()}
                disabled={isFetchingNextPage}
              >
                {isFetchingNextPage ? "로딩 중..." : "더 보기"}
              </Button>
            </div>
          )}
        </div>
      )}

      <AlertDialog
        open={deleteTarget !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>영구 삭제하시겠습니까?</AlertDialogTitle>
            <AlertDialogDescription>
              이 작업은 되돌릴 수 없습니다. 게시글이 완전히 삭제됩니다.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>취소</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              onClick={() => {
                if (deleteTarget) {
                  permanentDeleteMutation.mutate(deleteTarget);
                  setDeleteTarget(null);
                }
              }}
            >
              영구 삭제
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}

function TrashItem({
  post,
  onRestore,
  onPermanentDelete,
  isRestoring,
}: {
  post: TrashPost;
  onRestore: (id: string) => void;
  onPermanentDelete: (id: string) => void;
  isRestoring: boolean;
}) {
  return (
    <div className="border-b border-border px-4 py-3">
      <div className="flex items-start gap-3">
        <UserAvatar
          profileImageUrl={post.author.profileImageUrl}
          displayName={post.author.displayName || post.author.username}
          size="md"
        />
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <span className="text-[15px] font-bold truncate">
              {post.author.displayName || post.author.username}
            </span>
            <span className="text-sm text-muted-foreground truncate">
              @{post.author.username}
            </span>
          </div>
          <p className="mt-1 text-[15px] line-clamp-2 whitespace-pre-wrap break-words">
            {post.content}
          </p>
          <div className="mt-2 flex items-center justify-between">
            <span className="text-xs text-muted-foreground">
              삭제일: {formatRelativeTime(post.deletedAt)}
            </span>
            <div className="flex items-center gap-2">
              {post.canRestore ? (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onRestore(post.id)}
                  disabled={isRestoring}
                  className="h-8 gap-1.5"
                >
                  <RotateCcw size={14} />
                  복원
                </Button>
              ) : (
                <span className="text-xs text-muted-foreground">
                  복원 기간 만료
                </span>
              )}
              <Button
                variant="outline"
                size="sm"
                onClick={() => onPermanentDelete(post.id)}
                className="h-8 gap-1.5 text-destructive hover:bg-destructive/10 hover:text-destructive border-destructive/30"
              >
                <Trash2 size={14} />
                영구 삭제
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

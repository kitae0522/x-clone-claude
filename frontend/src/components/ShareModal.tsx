import { Link2, Mail } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface ShareModalProps {
  open: boolean;
  onClose: () => void;
  postId: string;
}

function getPostUrl(postId: string) {
  return `${window.location.origin}/post/${postId}`;
}

export default function ShareModal({ open, onClose, postId }: ShareModalProps) {
  const url = getPostUrl(postId);

  function handleCopyLink() {
    navigator.clipboard.writeText(url).then(() => {
      toast.success("링크가 복사되었습니다");
      onClose();
    });
  }

  const shareOptions = [
    {
      label: "이메일",
      icon: <Mail className="h-5 w-5" />,
      onClick: () => {
        window.open(
          `mailto:?body=${encodeURIComponent(url)}`,
        );
        onClose();
      },
    },
    {
      label: "링크 복사",
      icon: <Link2 className="h-5 w-5" />,
      onClick: handleCopyLink,
    },
  ];

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-xs">
        <DialogHeader>
          <DialogTitle>공유하기</DialogTitle>
        </DialogHeader>
        <div className="grid grid-cols-2 gap-4 py-2">
          {shareOptions.map(({ label, icon, onClick }) => (
            <button
              key={label}
              onClick={onClick}
              className="flex cursor-pointer flex-col items-center gap-2 rounded-xl border-none bg-transparent p-2 text-foreground transition-colors hover:bg-muted"
            >
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-muted">
                {icon}
              </div>
              <span className="text-[11px] text-muted-foreground">{label}</span>
            </button>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}

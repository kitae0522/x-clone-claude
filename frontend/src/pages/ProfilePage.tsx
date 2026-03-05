import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { ArrowLeft, CalendarDays } from "lucide-react";
import { useProfile } from "@/hooks/useProfile";
import { useAuth } from "@/hooks/useAuthContext";
import { useFollow, useUnfollow } from "@/hooks/useFollow";
import EditProfileModal from "@/components/EditProfileModal";
import FollowListModal from "@/components/FollowListModal";
import UserAvatar from "@/components/UserAvatar";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type Tab = "posts" | "replies" | "likes";

export default function ProfilePage() {
  const { handle } = useParams<{ handle: string }>();
  const navigate = useNavigate();
  const { user: currentUser } = useAuth();
  const { data: profile, isLoading, error } = useProfile(handle ?? "");
  const [showEditModal, setShowEditModal] = useState(false);
  const [followListType, setFollowListType] = useState<
    "followers" | "following" | null
  >(null);
  const [isHoveringFollow, setIsHoveringFollow] = useState(false);
  const [activeTab, setActiveTab] = useState<Tab>("posts");

  const follow = useFollow(handle ?? "");
  const unfollow = useUnfollow(handle ?? "");

  const isOwner = currentUser?.username === profile?.username;

  if (isLoading) {
    return (
      <div className="flex justify-center py-8">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  if (error || !profile) {
    return (
      <>
        <header className="sticky top-0 z-10 flex items-center gap-4 border-b border-border bg-background/65 px-4 py-2 backdrop-blur-xl">
          <button
            onClick={() => navigate(-1)}
            className="cursor-pointer rounded-full border-none bg-transparent p-2 text-foreground transition-colors hover:bg-foreground/10"
          >
            <ArrowLeft className="h-5 w-5" />
          </button>
          <span className="text-xl font-bold">프로필</span>
        </header>
        <p className="px-4 py-8 text-center text-destructive">
          {error?.message ?? "사용자를 찾을 수 없습니다."}
        </p>
      </>
    );
  }

  const joinedDate = new Date(profile.createdAt).toLocaleDateString("ko-KR", {
    year: "numeric",
    month: "long",
  });

  function handleFollowClick() {
    if (profile?.isFollowing) {
      unfollow.mutate();
    } else {
      follow.mutate();
    }
  }

  const tabs: { key: Tab; label: string }[] = [
    { key: "posts", label: "게시물" },
    { key: "replies", label: "답글" },
    { key: "likes", label: "마음에 들어요" },
  ];

  return (
    <>
      {/* Header */}
      <header className="sticky top-0 z-10 flex items-center gap-4 border-b border-border bg-background/65 px-4 py-2 backdrop-blur-xl">
        <button
          onClick={() => navigate(-1)}
          className="cursor-pointer rounded-full border-none bg-transparent p-2 text-foreground transition-colors hover:bg-foreground/10"
        >
          <ArrowLeft className="h-5 w-5" />
        </button>
        <div>
          <div className="text-xl font-bold leading-tight">
            {profile.displayName}
          </div>
          <div className="text-[13px] text-muted-foreground">0 posts</div>
        </div>
      </header>

      {/* Banner */}
      {profile.headerImageUrl ? (
        <img
          src={profile.headerImageUrl}
          alt="헤더 이미지"
          className="h-[200px] w-full object-cover"
        />
      ) : (
        <div className="h-[200px] w-full bg-muted-foreground/20" />
      )}

      {/* Profile Info */}
      <div className="px-4">
        <div className="-mt-[66px] flex items-end justify-between">
          <UserAvatar
            profileImageUrl={profile.profileImageUrl}
            displayName={profile.displayName}
            size="2xl"
            className="border-4 border-background"
          />
          {isOwner ? (
            <Button
              onClick={() => setShowEditModal(true)}
              variant="outline"
              size="sm"
              className="cursor-pointer rounded-full"
            >
              프로필 수정
            </Button>
          ) : currentUser ? (
            <Button
              onClick={handleFollowClick}
              onMouseEnter={() => setIsHoveringFollow(true)}
              onMouseLeave={() => setIsHoveringFollow(false)}
              variant={
                profile.isFollowing
                  ? isHoveringFollow
                    ? "follow-danger"
                    : "follow-active"
                  : "follow"
              }
              size="sm"
              className="min-w-[100px] cursor-pointer"
              disabled={follow.isPending || unfollow.isPending}
            >
              {profile.isFollowing
                ? isHoveringFollow
                  ? "언팔로우"
                  : "팔로잉"
                : "팔로우"}
            </Button>
          ) : null}
        </div>

        <div className="mt-3">
          <div className="text-xl font-bold">{profile.displayName}</div>
          <div className="text-[15px] text-muted-foreground">
            @{profile.username}
          </div>
        </div>

        {profile.bio && (
          <p className="mt-3 whitespace-pre-wrap text-[15px] leading-relaxed">
            {profile.bio}
          </p>
        )}

        <div className="mt-3 flex items-center gap-1 text-sm text-muted-foreground">
          <CalendarDays className="h-4 w-4" />
          <span>{joinedDate} 가입</span>
        </div>

        <div className="mt-3 flex gap-5">
          <span
            className="cursor-pointer text-sm text-muted-foreground hover:underline"
            onClick={() => setFollowListType("following")}
          >
            <strong className="text-foreground">
              {profile.followingCount}
            </strong>{" "}
            팔로잉
          </span>
          <span
            className="cursor-pointer text-sm text-muted-foreground hover:underline"
            onClick={() => setFollowListType("followers")}
          >
            <strong className="text-foreground">
              {profile.followersCount}
            </strong>{" "}
            팔로워
          </span>
        </div>
      </div>

      {/* Tabs */}
      <div className="mt-3 flex border-b border-border">
        {tabs.map(({ key, label }) => (
          <button
            key={key}
            onClick={() => setActiveTab(key)}
            className={cn(
              "relative flex-1 cursor-pointer border-none bg-transparent py-4 text-center text-[15px] font-medium transition-colors hover:bg-foreground/5",
              activeTab === key
                ? "font-bold text-foreground"
                : "text-muted-foreground",
            )}
          >
            {label}
            {activeTab === key && (
              <div className="absolute bottom-0 left-1/2 h-1 w-14 -translate-x-1/2 rounded-full bg-primary" />
            )}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div className="py-8 text-center text-sm text-muted-foreground">
        {activeTab === "posts" && "아직 게시물이 없습니다."}
        {activeTab === "replies" && "아직 답글이 없습니다."}
        {activeTab === "likes" && "아직 좋아요한 게시물이 없습니다."}
      </div>

      {showEditModal && currentUser && (
        <EditProfileModal
          user={currentUser}
          onClose={() => setShowEditModal(false)}
        />
      )}

      {followListType && handle && (
        <FollowListModal
          handle={handle}
          type={followListType}
          onClose={() => setFollowListType(null)}
        />
      )}
    </>
  );
}

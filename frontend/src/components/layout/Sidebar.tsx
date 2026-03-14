import { NavLink, useNavigate } from "react-router-dom";
import { Home, User, LogOut, Feather, Settings } from "lucide-react";
import { useAuth } from "@/hooks/useAuthContext";
import UserAvatar from "@/components/UserAvatar";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const navItems = [
  { to: "/", icon: Home, label: "홈" },
  { to: "/settings", icon: Settings, label: "설정" },
];

export default function Sidebar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  return (
    <aside className="flex h-dvh flex-col items-end border-r border-border px-2 py-3 xl:px-4">
      <div className="flex flex-1 flex-col gap-1 xl:w-full">
        {/* Logo */}
        <div className="mb-2 flex h-12 w-12 items-center justify-center rounded-full transition-colors hover:bg-primary/10 xl:ml-1">
          <Feather className="h-7 w-7 text-primary" />
        </div>

        {/* Nav Links */}
        {navItems.map(({ to, icon: Icon, label }) => (
          <NavLink
            key={to}
            to={to}
            className={({ isActive }) =>
              cn(
                "group flex items-center gap-4 rounded-full px-3 py-3 text-[20px] transition-colors hover:bg-foreground/10 xl:pr-6",
                isActive ? "font-bold text-foreground" : "text-foreground/80",
              )
            }
          >
            <Icon className="h-[26px] w-[26px]" strokeWidth={1.75} />
            <span className="hidden xl:inline">{label}</span>
          </NavLink>
        ))}

        {/* Profile Link */}
        {user && (
          <NavLink
            to={`/${user.username}`}
            className={({ isActive }) =>
              cn(
                "group flex items-center gap-4 rounded-full px-3 py-3 text-[20px] transition-colors hover:bg-foreground/10 xl:pr-6",
                isActive ? "font-bold text-foreground" : "text-foreground/80",
              )
            }
          >
            <User className="h-[26px] w-[26px]" strokeWidth={1.75} />
            <span className="hidden xl:inline">프로필</span>
          </NavLink>
        )}

        {/* Post Button */}
        <Button
          onClick={() => navigate("/compose")}
          className="mt-3 h-12 w-12 rounded-full text-[17px] font-bold xl:w-full"
        >
          <Feather className="h-5 w-5 xl:hidden" />
          <span className="hidden xl:inline">게시하기</span>
        </Button>
      </div>

      {/* User Info & Logout */}
      {user && (
        <button
          onClick={logout}
          className="flex w-full cursor-pointer items-center gap-3 rounded-full border-none bg-transparent p-3 text-left transition-colors hover:bg-foreground/10"
        >
          <UserAvatar
            profileImageUrl={user.profileImageUrl}
            displayName={user.displayName}
            size="sm"
          />
          <div className="hidden min-w-0 flex-1 xl:block">
            <div className="truncate text-[15px] font-bold text-foreground">
              {user.displayName || user.username}
            </div>
            <div className="truncate text-[13px] text-muted-foreground">
              @{user.username}
            </div>
          </div>
          <LogOut className="hidden h-4 w-4 text-muted-foreground xl:block" />
        </button>
      )}
    </aside>
  );
}

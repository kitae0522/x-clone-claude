import { NavLink } from "react-router-dom";
import { Home, Feather, User } from "lucide-react";
import { useAuth } from "@/hooks/useAuthContext";
import { cn } from "@/lib/utils";

export default function MobileNav() {
  const { user } = useAuth();

  return (
    <nav className="fixed inset-x-0 bottom-0 z-50 flex items-center justify-around border-t border-border bg-background/80 pb-safe backdrop-blur-xl md:hidden">
      <NavLink
        to="/"
        className={({ isActive }) =>
          cn(
            "flex flex-1 items-center justify-center py-3 transition-colors",
            isActive ? "text-foreground" : "text-muted-foreground",
          )
        }
      >
        <Home className="h-6 w-6" strokeWidth={1.75} />
      </NavLink>

      <NavLink
        to="/compose"
        className="flex flex-1 items-center justify-center py-3"
      >
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-md">
          <Feather className="h-5 w-5" strokeWidth={2} />
        </div>
      </NavLink>

      {user && (
        <NavLink
          to={`/${user.username}`}
          className={({ isActive }) =>
            cn(
              "flex flex-1 items-center justify-center py-3 transition-colors",
              isActive ? "text-foreground" : "text-muted-foreground",
            )
          }
        >
          <User className="h-6 w-6" strokeWidth={1.75} />
        </NavLink>
      )}
    </nav>
  );
}

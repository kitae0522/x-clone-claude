import { NavLink } from "react-router-dom";
import { Home, User } from "lucide-react";
import { useAuth } from "@/hooks/useAuthContext";
import { cn } from "@/lib/utils";

export default function MobileNav() {
  const { user } = useAuth();

  const items = [
    { to: "/", icon: Home },
    ...(user ? [{ to: `/${user.username}`, icon: User }] : []),
  ];

  return (
    <nav className="fixed inset-x-0 bottom-0 z-50 flex items-center justify-around border-t border-border bg-background/80 pb-safe backdrop-blur-xl md:hidden">
      {items.map(({ to, icon: Icon }) => (
        <NavLink
          key={to}
          to={to}
          className={({ isActive }) =>
            cn(
              "flex flex-1 items-center justify-center py-3 transition-colors",
              isActive ? "text-foreground" : "text-muted-foreground",
            )
          }
        >
          <Icon className="h-6 w-6" strokeWidth={1.75} />
        </NavLink>
      ))}
    </nav>
  );
}

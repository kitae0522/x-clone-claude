import { Outlet } from "react-router-dom";
import Sidebar from "./Sidebar";
import MobileNav from "./MobileNav";
import RightPanel from "./RightPanel";

export default function MainLayout() {
  return (
    <div className="mx-auto flex min-h-dvh max-w-[1280px] justify-center">
      {/* Left Sidebar - hidden on mobile */}
      <div className="hidden w-[68px] shrink-0 md:block xl:w-[275px]">
        <div className="sticky top-0">
          <Sidebar />
        </div>
      </div>

      {/* Main Content */}
      <main className="w-full max-w-[600px] border-r border-border pb-16 md:pb-0">
        <Outlet />
      </main>

      {/* Right Panel - hidden on tablet and mobile */}
      <div className="hidden w-[350px] shrink-0 lg:block">
        <RightPanel />
      </div>

      {/* Mobile Bottom Nav */}
      <MobileNav />
    </div>
  );
}

import type { ComponentProps } from "react";
import { cn } from "@/lib/utils";

type ContentHeaderProps = ComponentProps<"header">;

export function ContentHeader({ className, ...props }: ContentHeaderProps) {
  return (
    <header
      className={cn(
        "liquid-edge flex h-[52px] shrink-0 items-center justify-between border-b border-border bg-panel px-3 shadow-[var(--panel-shadow)] sm:px-4",
        className,
      )}
      {...props}
    />
  );
}

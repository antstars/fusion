import type { ComponentProps } from "react";
import { cn } from "@/lib/utils";

type ContentHeaderProps = ComponentProps<"header">;

export function ContentHeader({ className, ...props }: ContentHeaderProps) {
  return (
    <header
      className={cn(
        "flex h-12 shrink-0 items-center justify-between border-b border-border bg-panel px-3 sm:px-4",
        className,
      )}
      {...props}
    />
  );
}

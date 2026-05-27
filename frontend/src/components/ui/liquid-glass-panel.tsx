import type { CSSProperties, ReactNode } from "react";
import LiquidGlass from "liquid-glass-react";
import { cn } from "@/lib/utils";

interface LiquidGlassPanelProps {
  children: ReactNode;
  className?: string;
  contentClassName?: string;
  cornerRadius?: number;
  overLight?: boolean;
}

const liquidGlassStyle = {
  position: "absolute",
  inset: 0,
  top: "50%",
  left: "50%",
  width: "100%",
  height: "100%",
} satisfies CSSProperties;

export function LiquidGlassPanel({
  children,
  className,
  contentClassName,
  cornerRadius = 12,
  overLight = true,
}: LiquidGlassPanelProps) {
  return (
    <div className={cn("fusion-liquid-glass-panel", className)}>
      <LiquidGlass
        className="fusion-liquid-glass"
        style={liquidGlassStyle}
        padding="0"
        cornerRadius={cornerRadius}
        displacementScale={28}
        blurAmount={0.08}
        saturation={145}
        aberrationIntensity={0.8}
        elasticity={0.04}
        mode="standard"
        overLight={overLight}
      >
        <div className={cn("fusion-liquid-glass-content", contentClassName)}>
          {children}
        </div>
      </LiquidGlass>
    </div>
  );
}

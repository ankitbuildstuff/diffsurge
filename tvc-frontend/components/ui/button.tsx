import { cn } from "@/lib/utils";
import { type ButtonHTMLAttributes, forwardRef } from "react";

type Variant = "primary" | "secondary" | "ghost" | "outline";
type Size = "sm" | "md" | "lg";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
}

const variantStyles: Record<Variant, string> = {
  primary:
    "text-[var(--bg-primary)] active:scale-[0.98]",
  secondary:
    "border active:scale-[0.98]",
  ghost:
    "active:scale-[0.98]",
  outline:
    "border active:scale-[0.98]",
};

const sizeStyles: Record<Size, string> = {
  sm: "h-8 px-4 text-[13px] gap-1.5",
  md: "h-10 px-5 text-[13px] gap-2",
  lg: "h-11 px-6 text-[14px] gap-2",
};

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "primary", size = "md", style, children, ...props }, ref) => {
    const variantInlineStyles: Record<Variant, React.CSSProperties> = {
      primary: {
        backgroundColor: "var(--bg-dark)",
        color: "var(--bg-primary)",
      },
      secondary: {
        backgroundColor: "var(--bg-primary)",
        color: "var(--text-secondary)",
        borderColor: "var(--border-light)",
      },
      ghost: {
        backgroundColor: "transparent",
        color: "var(--text-muted)",
      },
      outline: {
        backgroundColor: "transparent",
        color: "var(--text-secondary)",
        borderColor: "var(--border-light)",
      },
    };

    return (
      <button
        ref={ref}
        className={cn(
          "inline-flex items-center justify-center rounded-full font-medium transition-all duration-150 cursor-pointer whitespace-nowrap select-none focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--accent-purple)] disabled:pointer-events-none disabled:opacity-50 hover:opacity-85",
          variantStyles[variant],
          sizeStyles[size],
          className,
        )}
        style={{
          ...variantInlineStyles[variant],
          ...style,
        }}
        {...props}
      >
        {children}
      </button>
    );
  },
);

Button.displayName = "Button";

export { Button, type ButtonProps };

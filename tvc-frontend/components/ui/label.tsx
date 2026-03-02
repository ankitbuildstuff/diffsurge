import { cn } from "@/lib/utils";
import { type LabelHTMLAttributes, forwardRef } from "react";

export interface LabelProps extends LabelHTMLAttributes<HTMLLabelElement> {
  required?: boolean;
  error?: boolean;
}

const Label = forwardRef<HTMLLabelElement, LabelProps>(
  ({ className, required, error, children, ...props }, ref) => {
    return (
      <label
        ref={ref}
        className={cn(
          "text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70",
          className,
        )}
        style={{
          color: error ? "var(--accent-orange)" : "var(--text-secondary)",
        }}
        {...props}
      >
        {children}
        {required && (
          <span style={{ marginLeft: 4, color: "var(--accent-orange)" }}>*</span>
        )}
      </label>
    );
  },
);
Label.displayName = "Label";

export { Label };

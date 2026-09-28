import type { ButtonHTMLAttributes, ReactNode } from "react";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "secondary" | "danger" | "ghost";
  icon?: ReactNode;
  loading?: boolean;
};

export function Button({
  variant = "secondary",
  icon,
  loading = false,
  className = "",
  children,
  disabled,
  ...props
}: ButtonProps) {
  return (
    <button
      className={`ui-button ui-button-${variant}${className ? ` ${className}` : ""}`}
      disabled={disabled || loading}
      {...props}
    >
      {loading ? <span className="ui-spinner" aria-hidden="true" /> : icon}
      {children}
    </button>
  );
}

type IconButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  label: string;
  children: ReactNode;
  tone?: "neutral" | "danger";
};

export function IconButton({ label, children, tone = "neutral", className = "", ...props }: IconButtonProps) {
  return (
    <button
      className={`ui-icon-button ui-icon-button-${tone}${className ? ` ${className}` : ""}`}
      aria-label={label}
      title={props.title ?? label}
      {...props}
    >
      {children}
    </button>
  );
}

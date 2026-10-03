import { WrapText } from "lucide-react";
import { Button } from "@/shared/ui";

type TableAppearanceProps = {
  wrapText: boolean;
  onWrapTextChange: (enabled: boolean) => void;
};

export function TableAppearance({ wrapText, onWrapTextChange }: TableAppearanceProps) {
  return (
    <div className="table-appearance" role="group" aria-label="Внешний вид таблицы">
      <Button
        type="button"
        icon={<WrapText size={16} />}
        aria-pressed={wrapText}
        onClick={() => onWrapTextChange(!wrapText)}
      >
        Перенос текста
      </Button>
    </div>
  );
}

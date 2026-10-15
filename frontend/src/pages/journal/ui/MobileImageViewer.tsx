import { useEffect, useRef, useState } from "react";
import { X, ZoomIn, ZoomOut } from "lucide-react";
import { attachmentsApi, type Attachment } from "@/entities/attachment/api/attachments";
import { Button, IconButton } from "@/shared/ui";

type Props = {
  images: Attachment[];
  onClose: () => void;
  onCompare: (image: Attachment) => void;
};

export function MobileImageViewer({ images, onClose, onCompare }: Props) {
  const ref = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const dialog = ref.current;
    const overflow = document.body.style.overflow;
    dialog?.showModal();
    document.body.style.overflow = "hidden";
    return () => {
      dialog?.close();
      document.body.style.overflow = overflow;
    };
  }, []);

  return (
    <dialog ref={ref} className="mobile-image-viewer" aria-labelledby="image-viewer-title" onCancel={onClose}>
      <header className="mobile-image-viewer-header">
        <h2 id="image-viewer-title">{images.length > 1 ? "Сравнение скриншотов" : "Просмотр скриншота"}</h2>
        <IconButton label="Закрыть просмотр" onClick={onClose}><X size={20} /></IconButton>
      </header>
      <div className={`mobile-image-panels${images.length > 1 ? " comparing" : ""}`}>
        {images.map((image) => <ImagePanel key={image.id} image={image} />)}
      </div>
      <footer className="mobile-image-viewer-footer">
        <Button onClick={() => onCompare(images[0])}>{images.length > 1 ? "Выбрать другой скриншот" : "Сравнить"}</Button>
      </footer>
    </dialog>
  );
}

function ImagePanel({ image }: { image: Attachment }) {
  const [zoom, setZoom] = useState(1);
  const pinch = useRef<{ distance: number; zoom: number } | null>(null);
  return (
    <section className="mobile-image-panel" aria-label={image.name}>
      <div className="mobile-image-panel-tools">
        <span title={image.name}>{image.name}</span>
        <IconButton label={`Уменьшить ${image.name}`} disabled={zoom <= 1} onClick={() => setZoom(Math.max(1, zoom - 0.5))}><ZoomOut size={18} /></IconButton>
        <IconButton label={`Увеличить ${image.name}`} disabled={zoom >= 4} onClick={() => setZoom(Math.min(4, zoom + 0.5))}><ZoomIn size={18} /></IconButton>
      </div>
      <div
        className="mobile-image-scroll"
        onTouchStart={(event) => {
          if (event.touches.length !== 2) return;
          const [first, second] = Array.from(event.touches);
          pinch.current = { distance: Math.hypot(first.clientX - second.clientX, first.clientY - second.clientY), zoom };
        }}
        onTouchMove={(event) => {
          if (event.touches.length !== 2 || !pinch.current || !pinch.current.distance) return;
          const [first, second] = Array.from(event.touches);
          const distance = Math.hypot(first.clientX - second.clientX, first.clientY - second.clientY);
          setZoom(Math.min(4, Math.max(1, pinch.current.zoom * distance / pinch.current.distance)));
        }}
        onTouchEnd={() => { pinch.current = null; }}
        onTouchCancel={() => { pinch.current = null; }}
      >
        <div className="mobile-image-canvas" style={{ width: `${zoom * 100}%`, height: `${zoom * 100}%` }}>
          <img src={attachmentsApi.contentUrl(image.id)} alt={image.name} />
        </div>
      </div>
    </section>
  );
}

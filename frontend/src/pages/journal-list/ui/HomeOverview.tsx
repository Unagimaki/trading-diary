import { ArrowRight, BookOpen, ChartNoAxesCombined, Mail, Plus, SlidersHorizontal } from "lucide-react";
import { Link } from "react-router-dom";
import type { Journal } from "@/entities/journal/api/journals";
import { Button } from "@/shared/ui";

export function HomeOverview({ loading, journal, onCreate }: { loading: boolean; journal?: Journal; onCreate: () => void }) {
  return (
    <>
      <section className="home-intro" aria-labelledby="home-title">
        <span className="page-eyebrow">Ваш торговый дневник</span>
        <h1 id="home-title">От наблюдений<br />к выводам.</h1>
        <p>Сохраняйте сделки, скриншоты и наблюдения. Анализируйте результаты и проверяйте торговые гипотезы на своей истории.</p>
        <div className="home-primary-action">
          {journal ? <>
            <Link className="ui-button ui-button-primary" to={`/journals/${journal.id}`}>Продолжить работу <ArrowRight size={16} /></Link>
            <span className="home-last-journal" title={journal.name}>{journal.name}</span>
          </> : <Button variant="primary" loading={loading} disabled={loading} icon={<Plus size={16} />} onClick={onCreate}>Создать первый журнал</Button>}
        </div>
        <span className="home-intro-mark" aria-hidden="true"><BookOpen strokeWidth={1} /></span>
      </section>
      <section className="home-start" aria-labelledby="quick-start-title">
        <div className="home-section-label"><h2 id="quick-start-title">Быстрый старт</h2><span>Три шага к своей истории торговли</span></div>
        <ol className="home-steps">
          <li><span className="home-step-number">01</span><BookOpen size={18} /><h3>Создайте журнал</h3><p>Отдельно для стратегии, рынка или новой гипотезы.</p></li>
          <li><span className="home-step-number">02</span><SlidersHorizontal size={18} /><h3>Настройте колонки</h3><p>Выберите данные, которые помогут проверить вашу идею.</p></li>
          <li><span className="home-step-number">03</span><ChartNoAxesCombined size={18} /><h3>Находите закономерности</h3><p>Добавляйте сделки и скриншоты, смотрите аналитику.</p></li>
        </ol>
      </section>
    </>
  );
}

export function HomeFeedback() {
  return (
    <footer className="home-feedback">
      <span className="home-feedback-icon"><Mail size={19} /></span>
      <div><h2>Помогите сделать дневник лучше</h2><p>Нашли ошибку или не хватает возможности? Напишите мне.</p></div>
      <a className="ui-button ui-button-secondary" href="mailto:m.yark0v.2000@gmail.com?subject=Trade%20Diary%20%E2%80%94%20%D0%BE%D0%B1%D1%80%D0%B0%D1%82%D0%BD%D0%B0%D1%8F%20%D1%81%D0%B2%D1%8F%D0%B7%D1%8C">Напишите мне <ArrowRight size={14} /></a>
    </footer>
  );
}

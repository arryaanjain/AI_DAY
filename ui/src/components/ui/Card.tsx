import React from 'react';
import { Link } from 'react-router-dom';
import { ArrowRight } from 'lucide-react';

interface CardProps {
  children: React.ReactNode;
  className?: string;
  hoverEffect?: boolean;
}

export const Card: React.FC<CardProps> = ({
  children,
  className = '',
  hoverEffect = false,
}) => {
  return (
    <div
      className={`rounded-[2rem] border border-white/10 bg-white/5 p-6 backdrop-blur-xl transition-all duration-300 shadow-xl shadow-black/20 ${
        hoverEffect
          ? 'hover:-translate-y-1 hover:border-cyan-300/40 hover:bg-white/10 hover:shadow-2xl hover:shadow-cyan-950/30'
          : ''
      } ${className}`}
    >
      {children}
    </div>
  );
};

interface StatTileProps {
  label: string;
  value: string;
  hint: string;
  icon?: React.ReactNode;
}

export const StatTile: React.FC<StatTileProps> = ({ label, value, hint, icon }) => {
  return (
    <Card className="p-6">
      <div className="flex items-center justify-between">
        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">
          {label}
        </p>
        {icon && <div className="text-cyan-300">{icon}</div>}
      </div>
      <p className="mt-3 text-3xl font-bold tracking-tight text-white">{value}</p>
      <p className="mt-1 text-xs text-slate-400">{hint}</p>
    </Card>
  );
};

interface HeroCardProps {
  title: string;
  description: string;
  action: string;
  href: string;
  icon?: React.ReactNode;
  badge?: string;
}

export const HeroCard: React.FC<HeroCardProps> = ({
  title,
  description,
  action,
  href,
  icon,
  badge,
}) => {
  return (
    <Link to={href} className="group block">
      <Card hoverEffect className="h-full p-8 flex flex-col justify-between">
        <div>
          <div className="flex items-center justify-between gap-4">
            {icon && (
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-cyan-300/10 text-cyan-300 border border-cyan-300/20 group-hover:scale-110 transition-transform">
                {icon}
              </div>
            )}
            {badge && (
              <span className="rounded-full border border-cyan-300/30 bg-cyan-300/10 px-3 py-1 text-[10px] font-semibold tracking-wider text-cyan-300 uppercase">
                {badge}
              </span>
            )}
          </div>
          <h3 className="mt-6 text-2xl font-bold text-white group-hover:text-cyan-300 transition-colors">
            {title}
          </h3>
          <p className="mt-3 text-sm leading-relaxed text-slate-300">
            {description}
          </p>
        </div>
        <div className="mt-8 flex items-center gap-2 text-sm font-semibold text-cyan-300 group-hover:translate-x-1 transition-transform">
          <span>{action}</span>
          <ArrowRight className="h-4 w-4" />
        </div>
      </Card>
    </Link>
  );
};

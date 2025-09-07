'use client'

import React from 'react'
import { cn } from '@/lib/utils'

interface NavigationItem {
  id: string
  name: string
  active: boolean
}

interface GlassNavigationProps {
  items: NavigationItem[]
  onItemClick: (itemId: string, itemName: string) => void
  className?: string
}

export function GlassNavigation({ items, onItemClick, className }: GlassNavigationProps) {
  return (
    <div className={cn(
      "flex space-x-1 p-1 rounded-2xl",
      "bg-white/30 border border-white/40",
      "shadow-xl shadow-black/10",
      "transition-all duration-300 ease-in-out",
      "relative overflow-hidden",
      className
    )}
    style={{
      backdropFilter: 'blur(12px)',
      WebkitBackdropFilter: 'blur(12px)',
      background: 'rgba(255, 255, 255, 0.25)'
    }}>
      {items.map((item) => (
        <button
          key={item.id}
          className={cn(
            "px-6 py-3 rounded-xl text-sm font-medium transition-all duration-300 ease-in-out",
            "relative overflow-hidden group",
            "hover:scale-105 active:scale-95",
            item.active
              ? [
                  "bg-white/95 text-gray-900 shadow-lg shadow-black/15",
                  "border border-white/60",
                  "before:absolute before:inset-0 before:bg-gradient-to-r before:from-blue-500/15 before:to-purple-500/15",
                  "before:opacity-100 before:transition-opacity before:duration-300"
                ]
              : [
                  "text-gray-700 hover:text-gray-900",
                  "hover:bg-white/50 hover:border hover:border-white/40",
                  "hover:shadow-md hover:shadow-black/10",
                  "before:absolute before:inset-0 before:bg-gradient-to-r before:from-blue-500/8 before:to-purple-500/8",
                  "before:opacity-0 before:transition-opacity before:duration-300",
                  "hover:before:opacity-100"
                ]
          )}
          onClick={() => onItemClick(item.id, item.name)}
          style={item.active ? {
            backdropFilter: 'blur(8px)',
            WebkitBackdropFilter: 'blur(8px)'
          } : {}}
        >
          <span className="relative z-10">{item.name}</span>
          
          {/* 活跃状态的光晕效果 */}
          {item.active && (
            <div className="absolute inset-0 rounded-xl bg-gradient-to-r from-blue-400/20 to-purple-400/20 blur-sm" />
          )}
          
          {/* 悬停时的微光效果 */}
          <div className="absolute inset-0 rounded-xl opacity-0 group-hover:opacity-100 transition-opacity duration-300 bg-gradient-to-r from-blue-200/20 to-sky-200/10" />
        </button>
      ))}
    </div>
  )
}

// 预设样式变体
export const GlassNavigationVariants = {
  default: "bg-white/20 backdrop-blur-md border border-white/30",
  dark: "bg-black/20 backdrop-blur-md border border-white/10",
  colorful: "bg-gradient-to-r from-blue-500/20 to-purple-500/20 backdrop-blur-md border border-white/30",
  minimal: "bg-white/10 backdrop-blur-sm border border-white/20"
}

// 带有图标支持的增强版本
interface NavigationItemWithIcon extends NavigationItem {
  icon?: React.ReactNode
}

interface GlassNavigationWithIconsProps {
  items: NavigationItemWithIcon[]
  onItemClick: (itemId: string, itemName: string) => void
  className?: string
  showIcons?: boolean
}

export function GlassNavigationWithIcons({ 
  items, 
  onItemClick, 
  className, 
  showIcons = false 
}: GlassNavigationWithIconsProps) {
  return (
    <div className={cn(
      "flex space-x-1 p-1 rounded-2xl",
      "bg-white/20 backdrop-blur-md border border-white/30",
      "shadow-lg shadow-black/5",
      "transition-all duration-300 ease-in-out",
      className
    )}>
      {items.map((item) => (
        <button
          key={item.id}
          className={cn(
            "px-6 py-3 rounded-xl text-sm font-medium transition-all duration-300 ease-in-out",
            "relative overflow-hidden group flex items-center space-x-2",
            "hover:scale-105 active:scale-95",
            item.active
              ? [
                  "bg-white/90 text-gray-900 shadow-lg shadow-black/10",
                  "backdrop-blur-sm border border-white/50",
                  "before:absolute before:inset-0 before:bg-gradient-to-r before:from-blue-500/10 before:to-purple-500/10",
                  "before:opacity-100 before:transition-opacity before:duration-300"
                ]
              : [
                  "text-gray-700 hover:text-gray-900",
                  "hover:bg-white/40 hover:backdrop-blur-sm",
                  "hover:shadow-md hover:shadow-black/5",
                  "before:absolute before:inset-0 before:bg-gradient-to-r before:from-blue-500/5 before:to-purple-500/5",
                  "before:opacity-0 before:transition-opacity before:duration-300",
                  "hover:before:opacity-100"
                ]
          )}
          onClick={() => onItemClick(item.id, item.name)}
        >
          {showIcons && item.icon && (
            <span className="relative z-10 w-4 h-4">{item.icon}</span>
          )}
          <span className="relative z-10">{item.name}</span>
          
          {/* 活跃状态的光晕效果 */}
          {item.active && (
            <div className="absolute inset-0 rounded-xl bg-gradient-to-r from-blue-400/20 to-purple-400/20 blur-sm" />
          )}
          
          {/* 悬停时的微光效果 */}
          <div className="absolute inset-0 rounded-xl opacity-0 group-hover:opacity-100 transition-opacity duration-300 bg-gradient-to-r from-blue-200/20 to-sky-200/10" />
        </button>
      ))}
    </div>
  )
}
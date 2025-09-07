'use client'

import React, { useState, useRef, useEffect } from 'react'
import { cn } from '@/lib/utils'

interface DropdownMenuProps {
  children: React.ReactNode
  className?: string
}

interface DropdownMenuTriggerProps {
  children: React.ReactNode
  className?: string
  asChild?: boolean
}

interface DropdownMenuContentProps {
  children: React.ReactNode
  className?: string
  align?: 'start' | 'center' | 'end'
  side?: 'top' | 'right' | 'bottom' | 'left'
}

interface DropdownMenuItemProps {
  children: React.ReactNode
  className?: string
  onClick?: () => void
  disabled?: boolean
}

interface DropdownMenuSeparatorProps {
  className?: string
}

const DropdownMenuContext = React.createContext<{
  open: boolean
  setOpen: (open: boolean) => void
}>({ open: false, setOpen: () => {} })

const DropdownMenu = ({ children, className }: DropdownMenuProps) => {
  const [open, setOpen] = useState(false)
  
  return (
    <DropdownMenuContext.Provider value={{ open, setOpen }}>
      <div className={cn('relative inline-block text-left', className)}>
        {children}
      </div>
    </DropdownMenuContext.Provider>
  )
}

const DropdownMenuTrigger = ({ children, className, asChild }: DropdownMenuTriggerProps) => {
  const { open, setOpen } = React.useContext(DropdownMenuContext)
  
  const handleClick = () => {
    setOpen(!open)
  }
  
  if (asChild && React.isValidElement(children)) {
    const childElement = children as React.ReactElement<{ onClick?: () => void; className?: string }>
    return React.cloneElement(childElement, {
      onClick: handleClick,
      className: cn(childElement.props.className, className)
    })
  }
  
  return (
    <button
      onClick={handleClick}
      className={cn('inline-flex items-center justify-center', className)}
    >
      {children}
    </button>
  )
}

const DropdownMenuContent = ({ children, className, align = 'end', side = 'bottom' }: DropdownMenuContentProps) => {
  const { open, setOpen } = React.useContext(DropdownMenuContext)
  const contentRef = useRef<HTMLDivElement>(null)
  
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (contentRef.current && !contentRef.current.contains(event.target as Node)) {
        setOpen(false)
      }
    }
    
    if (open) {
      document.addEventListener('mousedown', handleClickOutside)
    }
    
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [open, setOpen])
  
  if (!open) return null
  
  const alignmentClasses = {
    start: 'left-0',
    center: 'left-1/2 transform -translate-x-1/2',
    end: 'right-0'
  }
  
  const sideClasses = {
    top: 'bottom-full mb-1',
    right: 'left-full ml-1 top-0',
    bottom: 'top-full mt-1',
    left: 'right-full mr-1 top-0'
  }
  
  return (
    <div
      ref={contentRef}
      className={cn(
        'absolute z-50 min-w-[8rem] overflow-hidden rounded-md border border-gray-700 bg-gray-800 p-1 text-gray-200 shadow-md',
        alignmentClasses[align],
        sideClasses[side],
        className
      )}
    >
      {children}
    </div>
  )
}

const DropdownMenuItem = ({ children, className, onClick, disabled }: DropdownMenuItemProps) => {
  const { setOpen } = React.useContext(DropdownMenuContext)
  
  const handleClick = () => {
    if (!disabled && onClick) {
      onClick()
      setOpen(false)
    }
  }
  
  return (
    <div
      onClick={handleClick}
      className={cn(
        'relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors',
        disabled
          ? 'pointer-events-none opacity-50'
          : 'hover:bg-gray-700 hover:text-white focus:bg-gray-700 focus:text-white',
        className
      )}
    >
      {children}
    </div>
  )
}

const DropdownMenuSeparator = ({ className }: DropdownMenuSeparatorProps) => {
  return (
    <div className={cn('-mx-1 my-1 h-px bg-gray-700', className)} />
  )
}

export {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator
}
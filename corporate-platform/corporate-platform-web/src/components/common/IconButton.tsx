'use client'

import { ReactNode } from 'react'

interface IconButtonProps {
  children: ReactNode
  label: string
  onClick?: () => void
  className?: string
  disabled?: boolean
  type?: 'button' | 'submit' | 'reset'
  'aria-expanded'?: boolean
  'aria-controls'?: string
  'aria-haspopup'?: boolean | 'menu' | 'listbox' | 'tree' | 'grid' | 'dialog'
  'aria-describedby'?: string
  'aria-pressed'?: boolean
}

export function IconButton({
  children,
  label,
  onClick,
  className = '',
  disabled = false,
  type = 'button',
  'aria-expanded': ariaExpanded,
  'aria-controls': ariaControls,
  'aria-haspopup': ariaHaspopup,
  'aria-describedby': ariaDescribedby,
  'aria-pressed': ariaPressed,
}: IconButtonProps) {
  return (
    <button
      type={type}
      onClick={onClick}
      className={className}
      disabled={disabled}
      aria-label={label}
      aria-expanded={ariaExpanded}
      aria-controls={ariaControls}
      aria-haspopup={ariaHaspopup}
      aria-describedby={ariaDescribedby}
      aria-pressed={ariaPressed}
    >
      {children}
    </button>
  )
}
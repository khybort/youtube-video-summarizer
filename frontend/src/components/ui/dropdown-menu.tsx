import * as React from "react"
import { cn } from "@/lib/utils"

interface DropdownMenuContextValue {
  open: boolean
  setOpen: (open: boolean) => void
}

const DropdownMenuContext = React.createContext<DropdownMenuContextValue | undefined>(undefined)

export function DropdownMenu({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = React.useState(false)

  React.useEffect(() => {
    if (open) {
      const handleClickOutside = (e: MouseEvent) => {
        const target = e.target as HTMLElement
        if (!target.closest('[data-dropdown-menu]')) {
          setOpen(false)
        }
      }
      document.addEventListener('click', handleClickOutside)
      return () => document.removeEventListener('click', handleClickOutside)
    }
  }, [open])

  return (
    <DropdownMenuContext.Provider value={{ open, setOpen }}>
      <div className="relative" data-dropdown-menu>
        {children}
      </div>
    </DropdownMenuContext.Provider>
  )
}

export function DropdownMenuTrigger({
  asChild,
  children,
}: {
  asChild?: boolean
  children: React.ReactNode
}) {
  const context = React.useContext(DropdownMenuContext)
  if (!context) throw new Error("DropdownMenuTrigger must be used within DropdownMenu")

  if (asChild && React.isValidElement(children)) {
    return React.cloneElement(children, {
      onClick: () => context.setOpen(!context.open),
    } as any)
  }

  return (
    <button onClick={() => context.setOpen(!context.open)}>
      {children}
    </button>
  )
}

export function DropdownMenuContent({
  align = "start",
  children,
}: {
  align?: "start" | "end" | "center"
  children: React.ReactNode
}) {
  const context = React.useContext(DropdownMenuContext)
  if (!context) throw new Error("DropdownMenuContent must be used within DropdownMenu")

  if (!context.open) return null

  const alignClasses = {
    start: "left-0",
    end: "right-0",
    center: "left-1/2 -translate-x-1/2",
  }

  return (
    <div
      className={cn(
        "absolute top-full mt-1 z-50 min-w-[8rem] rounded-md border bg-popover p-1 text-popover-foreground shadow-md",
        alignClasses[align]
      )}
    >
      {children}
    </div>
  )
}

export function DropdownMenuItem({
  children,
  onClick,
  className,
  disabled,
}: {
  children: React.ReactNode
  onClick?: () => void
  className?: string
  disabled?: boolean
}) {
  const context = React.useContext(DropdownMenuContext)
  if (!context) throw new Error("DropdownMenuItem must be used within DropdownMenu")

  return (
    <button
      className={cn(
        "relative flex w-full cursor-pointer select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
        className
      )}
      onClick={() => {
        onClick?.()
        context.setOpen(false)
      }}
      disabled={disabled}
    >
      {children}
    </button>
  )
}


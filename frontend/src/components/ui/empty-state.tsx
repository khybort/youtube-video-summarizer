import { ReactNode } from 'react'
import { Card, CardContent } from './card'

interface EmptyStateProps {
  icon?: ReactNode
  title: string
  description?: string
  action?: ReactNode
}

export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <Card>
      <CardContent className="py-12 text-center">
        {icon && <div className="mb-4 flex justify-center">{icon}</div>}
        <h3 className="text-lg font-semibold mb-2">{title}</h3>
        {description && <p className="text-sm text-muted-foreground mb-4">{description}</p>}
        {action && <div>{action}</div>}
      </CardContent>
    </Card>
  )
}


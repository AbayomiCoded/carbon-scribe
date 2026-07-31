import { useCallback, useMemo } from 'react'
import { useTranslation } from 'next-i18next'

interface AccessibilityLabels {
  toggleSidebar: string
  toggleTheme: string
  openCart: string
  viewNotifications: string
  openSettings: string
  toggleNavigation: string
  searchCredits: string
  logout: string
  backToCart: string
  removeItem: string
  closeCart: string
}

export function useAccessibility() {
  const { t } = useTranslation('common')

  const labels = useMemo<AccessibilityLabels>(() => ({
    toggleSidebar: t('accessibility.toggleSidebar', 'Toggle sidebar'),
    toggleTheme: t('accessibility.toggleTheme', 'Toggle theme'),
    openCart: t('accessibility.openCart', 'Open cart ({count} items)'),
    viewNotifications: t('accessibility.viewNotifications', 'View notifications'),
    openSettings: t('accessibility.openSettings', 'Open settings'),
    toggleNavigation: t('accessibility.toggleNavigation', 'Toggle navigation menu'),
    searchCredits: t('accessibility.searchCredits', 'Search credits, projects, or analytics...'),
    logout: t('accessibility.logout', 'Logout'),
    backToCart: t('accessibility.backToCart', 'Back to cart'),
    removeItem: t('accessibility.removeItem', 'Remove {itemName}'),
    closeCart: t('accessibility.closeCart', 'Close cart'),
  }), [t])

  const getCartLabel = useCallback((count: number): string => {
    return labels.openCart.replace('{count}', count.toString())
  }, [labels.openCart])

  const getRemoveItemLabel = useCallback((itemName: string): string => {
    return labels.removeItem.replace('{itemName}', itemName)
  }, [labels.removeItem])

  return {
    labels,
    getCartLabel,
    getRemoveItemLabel,
  }
}
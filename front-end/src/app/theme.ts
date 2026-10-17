import type { ThemeConfig } from 'antd'

/**
 * YouPass design system — Modern Premium EdTech SaaS.
 * Single source of truth for colors, typography, radii and spacing.
 */
export const colors = {
  primary: '#2563EB',
  primaryHover: '#1D4ED8',
  background: '#F7F9FC',
  surface: '#FFFFFF',
  text: '#111827',
  textSecondary: '#64748B',
  border: '#E5EAF0',
  success: '#16A34A',
  warning: '#D97706',
  error: '#DC2626',
} as const

export const fontFamily =
  "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif"

export const themeConfig: ThemeConfig = {
  token: {
    colorPrimary: colors.primary,
    colorInfo: colors.primary,
    colorLink: colors.primary,
    colorLinkHover: colors.primaryHover,
    colorSuccess: colors.success,
    colorWarning: colors.warning,
    colorError: colors.error,

    colorTextBase: colors.text,
    colorText: colors.text,
    colorTextSecondary: colors.textSecondary,
    colorTextTertiary: '#94A3B8',
    colorTextQuaternary: '#C7CFDA',

    colorBorder: colors.border,
    colorBorderSecondary: colors.border,
    colorSplit: colors.border,

    colorBgLayout: colors.background,
    colorBgContainer: colors.surface,
    colorBgElevated: colors.surface,
    colorFillAlter: colors.background,
    colorFillSecondary: '#F3F6FA',

    fontFamily,
    fontSize: 15,
    lineHeight: 1.6,

    borderRadius: 8,
    borderRadiusLG: 12,
    borderRadiusSM: 6,

    controlHeight: 40,
    controlHeightLG: 44,
    controlHeightSM: 32,
  },
  components: {
    Layout: {
      headerBg: colors.surface,
      siderBg: colors.surface,
      bodyBg: colors.background,
      headerHeight: 64,
      headerPadding: '0 24px',
    },
    Card: {
      borderRadiusLG: 12,
      paddingLG: 24,
      headerFontSize: 16,
    },
    Menu: {
      itemHeight: 44,
      itemBorderRadius: 8,
      itemSelectedBg: '#EFF6FF',
      itemSelectedColor: colors.primary,
      itemColor: '#475569',
      itemHoverBg: '#F3F6FA',
      itemHoverColor: colors.text,
      itemMarginInline: 12,
      iconSize: 18,
    },
    Button: {
      borderRadius: 8,
      controlHeight: 40,
      controlHeightLG: 44,
      fontWeight: 500,
    },
    Input: { controlHeight: 40, controlHeightLG: 44, borderRadius: 8 },
    InputNumber: { controlHeight: 40, borderRadius: 8 },
    Select: { controlHeight: 40, borderRadius: 8 },
    Modal: { borderRadiusLG: 16, titleFontSize: 18 },
    Table: {
      headerBg: colors.background,
      headerColor: '#475569',
      headerSplitColor: 'transparent',
      rowHoverBg: '#F8FAFD',
      cellPaddingBlock: 14,
      borderColor: colors.border,
    },
    Segmented: { trackBg: '#F3F6FA', itemSelectedBg: colors.surface },
    Tabs: { titleFontSize: 15 },
  },
}

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { useTranslation } from 'react-i18next';

// Stub i18n config since the other agent hasn't created it yet.
vi.mock('@/i18n/config', async () => {
  const i18next = await import('i18next');
  const { initReactI18next } = await import('react-i18next');
  const instance = i18next.default.createInstance();
  await instance.use(initReactI18next).init({
    lng: 'en',
    resources: { en: { translation: { hello: 'Hello' } } },
    interpolation: { escapeValue: false },
  });
  return { default: instance };
});

import { I18nProvider } from '@/components/providers/I18nProvider';

describe('I18nProvider', () => {
  it('renders children', () => {
    render(
      <I18nProvider>
        <span data-testid="child">Content</span>
      </I18nProvider>,
    );
    expect(screen.getByTestId('child')).toHaveTextContent('Content');
  });

  it('provides i18n context to children', () => {
    function TranslatedChild() {
      const { t } = useTranslation();
      return <span data-testid="translated">{t('hello')}</span>;
    }

    render(
      <I18nProvider>
        <TranslatedChild />
      </I18nProvider>,
    );
    expect(screen.getByTestId('translated')).toHaveTextContent('Hello');
  });
});

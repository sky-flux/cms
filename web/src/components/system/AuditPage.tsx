import { useState, useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';

import { QueryProvider } from '@/components/providers/QueryProvider';
import { I18nProvider } from '@/components/providers/I18nProvider';
import { AuditTable } from './AuditTable';
import { auditApi } from '@/lib/system-api';
import { usePagination } from '@/hooks/use-pagination';

function AuditPageInner() {
  const { t } = useTranslation();
  const { page, perPage, setPage, resetPage } = usePagination();

  const [actionFilter, setActionFilter] = useState<string>('');
  const [resourceTypeFilter, setResourceTypeFilter] = useState<string>('');
  const [startDate, setStartDate] = useState<string>('');
  const [endDate, setEndDate] = useState<string>('');

  const queryParams = {
    page,
    per_page: perPage,
    ...(actionFilter && actionFilter !== 'all' ? { action: actionFilter } : {}),
    ...(resourceTypeFilter && resourceTypeFilter !== 'all' ? { resource_type: resourceTypeFilter } : {}),
    ...(startDate ? { start_date: startDate } : {}),
    ...(endDate ? { end_date: endDate } : {}),
  };

  const { data, isLoading } = useQuery({
    queryKey: ['audit-logs', queryParams],
    queryFn: () => auditApi.list(queryParams),
  });

  const handleActionFilter = useCallback(
    (action: string) => {
      setActionFilter(action);
      resetPage();
    },
    [resetPage],
  );

  const handleResourceTypeFilter = useCallback(
    (resourceType: string) => {
      setResourceTypeFilter(resourceType);
      resetPage();
    },
    [resetPage],
  );

  const handleStartDateChange = useCallback(
    (date: string) => {
      setStartDate(date);
      resetPage();
    },
    [resetPage],
  );

  const handleEndDateChange = useCallback(
    (date: string) => {
      setEndDate(date);
      resetPage();
    },
    [resetPage],
  );

  const logs = data?.data ?? [];
  const pagination = data?.pagination ?? {
    page: 1,
    per_page: perPage,
    total: 0,
    total_pages: 1,
  };

  return (
    <div className="p-6">
      <h1 className="mb-6 text-2xl font-bold">{t('system.audit.title')}</h1>
      <AuditTable
        logs={logs}
        pagination={pagination}
        loading={isLoading}
        onPageChange={setPage}
        onActionFilter={handleActionFilter}
        onResourceTypeFilter={handleResourceTypeFilter}
        onStartDateChange={handleStartDateChange}
        onEndDateChange={handleEndDateChange}
      />
    </div>
  );
}

export function AuditPage() {
  return (
    <QueryProvider>
      <I18nProvider>
        <AuditPageInner />
      </I18nProvider>
    </QueryProvider>
  );
}

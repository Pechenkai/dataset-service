import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import { useServices } from '../context/ServiceContext';
import { DatasetFilters } from '../services/datasetService';

export const useCatalogViewModel = () => {
  const { datasetService } = useServices();
  const [searchParams, setSearchParams] = useSearchParams();

  const filters = useMemo(() => datasetService.buildFiltersFromSearch(searchParams), [datasetService, searchParams]);

  const categoriesQuery = useQuery(['categories'], () => datasetService.listCategories());

  const datasetsQuery = useQuery(['datasets', filters], () => datasetService.listDatasets(filters), {
    keepPreviousData: true
  });

  const updateFilters = (next: Partial<DatasetFilters>) => {
    const merged = { ...filters, ...next, page: next.page ?? 1 };
    setSearchParams(datasetService.stringifyFilters(merged));
  };

  return {
    filters,
    categories: categoriesQuery.data?.items || [],
    datasets: datasetsQuery.data?.items || [],
    meta: datasetsQuery.data?.meta,
    isLoading: datasetsQuery.isLoading,
    isFetching: datasetsQuery.isFetching,
    updateFilters
  };
};

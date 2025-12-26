import { useEffect, useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import { useServices } from '../context/ServiceContext';
import { DatasetFilters } from '../services/datasetService';
import { Dataset } from '../api/types';
import { persistentStore } from '../utils/persistentStore';

export const useCatalogViewModel = () => {
    const { datasetService } = useServices();
    const [searchParams, setSearchParams] = useSearchParams();
    const [filtersHydrated, setFiltersHydrated] = useState(false);
    const searchKey = searchParams.toString();

    const filters = useMemo(
        () => datasetService.buildFiltersFromSearch(searchParams),
        [datasetService, searchParams]
    );

    const categoriesQuery = useQuery({
        queryKey: ['categories'],
        queryFn: () => datasetService.listCategories()
    });

    useEffect(() => {
        if (filtersHydrated) return;
        const cached = persistentStore.getSync('catalog.filters');
        if (!cached || searchKey) {
            setFiltersHydrated(true);
            return;
        }
        try {
            const parsed = JSON.parse(cached) as DatasetFilters;
            setSearchParams(datasetService.stringifyFilters({ ...filters, ...parsed }));
        } catch {
            setFiltersHydrated(true);
        }
    }, [datasetService, filters, filtersHydrated, searchKey, setSearchParams]);

    useEffect(() => {
        if (!filtersHydrated) {
            return;
        }
        persistentStore.set('catalog.filters', JSON.stringify(filters));
    }, [filters, filtersHydrated]);

    const datasetsQuery = useQuery({
        queryKey: ['datasets', searchKey],
        queryFn: async () => {
            const page = await datasetService.listDatasets(filters);

            const items = page.items ?? [];
            const needsEnrich = items.some((d: any) => !d.latest_version?.id);

            if (!needsEnrich) return page;

            const enriched: Dataset[] = await Promise.all(
                items.map(async (d: any) => {
                    try {
                        const latest = await datasetService.getLatestVersion(d.id);
                        return { ...d, latest_version: latest ?? undefined };
                    } catch {
                        return d;
                    }
                })
            );

            return { ...page, items: enriched };
        },
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

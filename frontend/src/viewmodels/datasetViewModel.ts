import { useMemo } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import { useServices } from '../context/ServiceContext';

const intParam = (sp: URLSearchParams, key: string, def: number) => {
  const v = Number(sp.get(key));
  return Number.isFinite(v) && v > 0 ? v : def;
};

export const useDatasetViewModel = (datasetId: number) => {
  const { datasetService, subscriptionService } = useServices();
  const queryClient = useQueryClient();
  const [sp, setSp] = useSearchParams();

  const enabled = Number.isFinite(datasetId);

  const vpage = intParam(sp, 'vpage', 1);
  const rpage = intParam(sp, 'rpage', 1);

  const datasetQuery = useQuery({
    queryKey: ['dataset', datasetId],
    queryFn: () => datasetService.getDataset(datasetId),
    enabled
  });

  const versionsQuery = useQuery({
    queryKey: ['versions', datasetId, vpage],
    queryFn: () => datasetService.listVersions(datasetId, vpage, 5),
    enabled,
    keepPreviousData: true
  });

  const reviewsQuery = useQuery({
    queryKey: ['reviews', datasetId, rpage],
    queryFn: () => datasetService.listReviews(datasetId, rpage, 10),
    enabled,
    keepPreviousData: true
  });

  const createReview = useMutation({
    mutationFn: (payload: { rating: number; text?: string }) =>
        datasetService.createReview(datasetId, payload.rating, payload.text),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['reviews', datasetId] });
    }
  });

  const subscribe = useMutation({
    mutationFn: (userId?: number) => subscriptionService.subscribe(datasetId, userId)
  });

  const actions = useMemo(() => {
    const setVersionsPage = (p: number) => {
      const n = new URLSearchParams(sp);
      n.set('vpage', String(p));
      setSp(n, { replace: true });
    };
    const setReviewsPage = (p: number) => {
      const n = new URLSearchParams(sp);
      n.set('rpage', String(p));
      setSp(n, { replace: true });
    };
    return { setVersionsPage, setReviewsPage };
  }, [sp, setSp]);

  return { vpage, rpage, datasetQuery, versionsQuery, reviewsQuery, createReview, subscribe, actions };
};

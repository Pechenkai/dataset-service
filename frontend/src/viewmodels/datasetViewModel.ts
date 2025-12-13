import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useServices } from '../context/ServiceContext';

export const useDatasetViewModel = (datasetId: number) => {
  const { datasetService, subscriptionService } = useServices();
  const queryClient = useQueryClient();
  const enabled = Number.isFinite(datasetId);

  const datasetQuery = useQuery(['dataset', datasetId], () => datasetService.getDataset(datasetId), {
    enabled
  });
  const versionsQuery = useQuery(['versions', datasetId], () => datasetService.listVersions(datasetId), {
    enabled
  });
  const reviewsQuery = useQuery(['reviews', datasetId], () => datasetService.listReviews(datasetId), {
    enabled
  });

  const createReview = useMutation(
    (payload: { rating: number; text?: string }) =>
      datasetService.createReview(datasetId, payload.rating, payload.text),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['reviews', datasetId]);
      }
    }
  );

  const subscribe = useMutation((userId?: number) => subscriptionService.subscribe(datasetId, userId));

  return { datasetQuery, versionsQuery, reviewsQuery, createReview, subscribe };
};

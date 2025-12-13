import React, { createContext, useContext, useMemo } from 'react';
import { ApiClient } from '../api/client';
import { DatasetService } from '../services/datasetService';
import { NotificationService } from '../services/notificationService';
import { SubscriptionService } from '../services/subscriptionService';

interface ServiceBag {
  datasetService: DatasetService;
  notificationService: NotificationService;
  subscriptionService: SubscriptionService;
}

export const ServiceContext = createContext<ServiceBag | undefined>(undefined);

export const ServiceProvider: React.FC<{ apiClient: ApiClient; children: React.ReactNode }> = ({
  apiClient,
  children
}) => {
  const services = useMemo<ServiceBag>(() => {
    return {
      datasetService: new DatasetService(apiClient),
      notificationService: new NotificationService(apiClient),
      subscriptionService: new SubscriptionService(apiClient)
    };
  }, [apiClient]);

  return <ServiceContext.Provider value={services}>{children}</ServiceContext.Provider>;
};

export const useServices = () => {
  const ctx = useContext(ServiceContext);
  if (!ctx) {
    throw new Error('useServices must be used inside ServiceProvider');
  }
  return ctx;
};

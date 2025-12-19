import React, {createContext, useContext, useMemo} from 'react';
import {ApiClient} from '../api/client';
import {AuthService} from '../services/authService';
import {DatasetService} from '../services/datasetService';
import {NotificationService} from '../services/notificationService';
import {SubscriptionService} from '../services/subscriptionService';
import {createUserService, UserService} from '../services/userService';

export interface ServiceBag {
    apiClient: ApiClient;
    authService: AuthService;
    datasetService: DatasetService;
    notificationService: NotificationService;
    subscriptionService: SubscriptionService;
    userService: UserService;
}

export const ServiceContext = createContext<ServiceBag | undefined>(undefined);

export const ServiceProvider: React.FC<{ apiClient: ApiClient; children: React.ReactNode }> = ({
                                                                                                   apiClient,
                                                                                                   children
                                                                                               }) => {
    const services = useMemo<ServiceBag>(() => {
        return {
            apiClient,
            authService: new AuthService(apiClient),
            datasetService: new DatasetService(apiClient),
            notificationService: new NotificationService(apiClient),
            subscriptionService: new SubscriptionService(apiClient),
            userService: createUserService(apiClient)
        };
    }, [apiClient]);

    return <ServiceContext.Provider value={services}>{children}</ServiceContext.Provider>;
};

export const useServices = (): ServiceBag => {
    const ctx = useContext(ServiceContext);
    if (!ctx) throw new Error('useServices must be used inside ServiceProvider');
    return ctx;
};

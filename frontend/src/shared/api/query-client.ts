import { MutationCache, QueryClient } from "@tanstack/react-query";
import { ApiError, errorMessage, errorRequestId } from "./api-error";
import { useNotificationStore } from "@/shared/model/notification-store";

export const queryClient = new QueryClient({
  mutationCache: new MutationCache({
    onError: (error) =>
      useNotificationStore
        .getState()
        .show(errorMessage(error), errorRequestId(error)),
  }),
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: (failureCount, error) =>
        !(error instanceof ApiError && error.status === 401) && failureCount < 1,
    },
  },
});

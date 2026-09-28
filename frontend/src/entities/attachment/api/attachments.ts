import { API_URL } from "@/shared/config/env";
import { apiRequest } from "@/shared/api/http";
import { responseError } from "@/shared/api/api-error";
export type Attachment = {
  id: string;
  name: string;
  mimeType: string;
  size: number;
  createdAt: string;
};
export const attachmentsApi = {
  contentUrl: (id: string) => `${API_URL}/attachments/${id}/content`,
  upload: async (
    journalId: string,
    rowId: string,
    columnId: string,
    file: File,
  ) => {
    const body = new FormData();
    body.append("file", file);
    const response = await fetch(
      `${API_URL}/journals/${journalId}/rows/${rowId}/cells/${columnId}/image`,
      { method: "POST", body, credentials: "include" },
    );
    if (!response.ok) throw await responseError(response);
    return response.json() as Promise<Attachment>;
  },
  remove: (journalId: string, rowId: string, columnId: string) =>
    apiRequest<void>(
      `/journals/${journalId}/rows/${rowId}/cells/${columnId}/image`,
      { method: "DELETE" },
    ),
};

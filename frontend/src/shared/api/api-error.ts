export type ApiErrorPayload = {
  code?: string;
  message?: string;
  requestId?: string;
};

const messages: Record<string, string> = {
  IMAGE_TOO_LARGE: "Изображение превышает допустимый размер 10 МБ",
  INVALID_IMAGE: "Выберите изображение JPEG, PNG или WebP",
  IMAGE_UNSUPPORTED_TYPE: "Поддерживаются только изображения JPEG, PNG и WebP",
  IMAGE_NOT_FOUND: "Изображение или ячейка больше не существует",
  AUTHENTICATION_REQUIRED: "Сессия завершена. Войдите снова",
  INVALID_CREDENTIALS: "Неверный email или пароль",
  EMAIL_ALREADY_REGISTERED: "Этот email уже зарегистрирован",
  INVALID_CELL_VALUE: "Значение не соответствует типу колонки",
  INVALID_JOURNAL_SETTINGS: "Проверьте депозит и процент риска",
  INTERNAL_ERROR: "Не удалось выполнить операцию. Повторите попытку",
};

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
    public readonly requestId?: string,
  ) {
    super(messages[code] ?? message ?? "Не удалось выполнить запрос");
    this.name = "ApiError";
  }
}

export async function responseError(response: Response): Promise<ApiError> {
  const payload = (await response.json().catch(() => ({}))) as ApiErrorPayload;
  return new ApiError(
    response.status,
    payload.code ?? "REQUEST_FAILED",
    payload.message ?? response.statusText,
    payload.requestId ?? response.headers.get("X-Request-ID") ?? undefined,
  );
}

export function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : "Неизвестная ошибка";
}

export function errorRequestId(error: unknown): string | undefined {
  return error instanceof ApiError ? error.requestId : undefined;
}

package httptransport

import (
	"github.com/gorilla/mux"
	"net/http"
)

type logOperation struct{ name, message string }

var logOperations = map[string]logOperation{
	"GET /health":                                                      {"health.check", "Проверка доступности выполнена"},
	"GET /metrics":                                                     {"metrics.read", "Метрики получены"},
	"POST /auth/register":                                              {"auth.register", "Пользователь зарегистрирован, сессия создана"},
	"POST /auth/login":                                                 {"auth.login", "Пользователь вошёл в систему"},
	"POST /auth/logout":                                                {"auth.logout", "Выход из системы выполнен"},
	"GET /auth/me":                                                     {"auth.current_user", "Текущий пользователь определён"},
	"GET /journals":                                                    {"journal.list", "Список журналов получен"},
	"POST /journals":                                                   {"journal.create", "Журнал создан"},
	"GET /journals/{id}":                                               {"journal.get", "Журнал получен"},
	"PATCH /journals/{id}":                                             {"journal.rename", "Журнал переименован"},
	"PATCH /journals/{id}/settings":                                    {"journal.settings.update", "Настройки расчётов сохранены"},
	"DELETE /journals/{id}":                                            {"journal.delete", "Журнал удалён"},
	"GET /journals/{journalID}/columns":                                {"column.list", "Колонки журнала получены"},
	"POST /journals/{journalID}/columns":                               {"column.create", "Колонка создана"},
	"PATCH /journals/{journalID}/columns/{id}":                         {"column.update", "Настройки или порядок колонки изменены"},
	"DELETE /journals/{journalID}/columns/{id}":                        {"column.delete", "Колонка удалена"},
	"GET /journals/{journalID}/rows":                                   {"row.list", "Наблюдения журнала получены"},
	"POST /journals/{journalID}/rows":                                  {"row.create", "Наблюдение создано"},
	"DELETE /journals/{journalID}/rows/{rowID}":                        {"row.delete", "Наблюдение удалено"},
	"PATCH /journals/{journalID}/rows/{rowID}/cells/{columnID}":        {"cell.update", "Значение ячейки сохранено"},
	"DELETE /journals/{journalID}/rows/{rowID}/cells/{columnID}":       {"cell.clear", "Значение ячейки очищено"},
	"POST /journals/{journalID}/rows/{rowID}/cells/{columnID}/image":   {"attachment.upload", "Скриншот загружен"},
	"DELETE /journals/{journalID}/rows/{rowID}/cells/{columnID}/image": {"attachment.delete", "Скриншот удалён"},
	"GET /attachments/{id}/content":                                    {"attachment.content", "Содержимое скриншота отправлено"},
	"GET /journals/{journalID}/analytics":                              {"analytics.get", "Аналитика журнала получена"},
}

func operationMessage(name string) string {
	for _, operation := range logOperations {
		if operation.name == name {
			return operation.message
		}
	}
	return "HTTP-запрос выполнен"
}

// The outer middleware also records 404, 405 and CORS requests; mux supplies route details here.
func logRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if info, ok := r.Context().Value(loggerKey).(*requestLogContext); ok {
			if route := mux.CurrentRoute(r); route != nil {
				info.route, _ = route.GetPathTemplate()
				if operation, found := logOperations[r.Method+" "+info.route]; found {
					info.operation = operation.name
				}
			}
			vars := mux.Vars(r)
			for _, field := range [][2]string{{"journalID", "journal_id"}, {"rowID", "row_id"}, {"columnID", "column_id"}} {
				if value := vars[field[0]]; value != "" {
					info.attrs = append(info.attrs, field[1], value)
				}
			}
			if id := vars["id"]; id != "" {
				key := "journal_id"
				if vars["journalID"] != "" {
					key = "column_id"
				}
				if info.operation == "attachment.content" {
					key = "attachment_id"
				}
				info.attrs = append(info.attrs, key, id)
			}
		}
		next.ServeHTTP(w, r)
	})
}

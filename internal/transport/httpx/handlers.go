package httpx

import (
	"net/http"
	"yandex-diplom/internal/auth"
	"yandex-diplom/internal/domain"
	"yandex-diplom/internal/gophermart"
)

func RegisterUser(svc gophermart.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		u, err := bindUserFromJSON(r)
		if err != nil {
			svc.WriteError(w, err)
			return
		}

		if len(u.Login) > 254 {
			svc.WriteError(w, domain.ErrInvalidPayload)
		}

		cookie, err := svc.Register(r.Context(), u)
		if err != nil {
			svc.WriteError(w, err)
			return
		}

		http.SetCookie(w, &cookie)

		w.WriteHeader(http.StatusOK)
	}
}

func LoginUser(svc gophermart.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		u, err := bindUserFromJSON(r)
		if err != nil {
			svc.WriteError(w, err)
			return
		}

		if len(u.Login) > 254 {
			svc.WriteError(w, domain.ErrInvalidPayload)
		}

		cookie, err := svc.Login(r.Context(), u)
		if err != nil {
			svc.WriteError(w, err)
			return
		}

		http.SetCookie(w, &cookie)

		w.WriteHeader(http.StatusOK)
	}
}

func CreateOrder(svc gophermart.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		o, err := bindOrderFromPlain(r)
		if err != nil {
			svc.WriteError(w, err)
			return
		}

		err = svc.PutOrder(r.Context(), user.Login, o)
		if err != nil {
			svc.WriteError(w, err)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func GetOrders(svc gophermart.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		orders, err := svc.GetOrders(r.Context(), user.Login)
		if err != nil {
			svc.WriteError(w, err)
			return
		}

		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if err := responseJSONFromOrders(w, orders); err != nil {
			svc.WriteError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetBalance(svc gophermart.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		balance, err := svc.GetBalance(r.Context(), *user)
		if err != nil {
			svc.WriteError(w, err)
			return
		}

		if err := responseJSONBalance(w, balance); err != nil {
			svc.WriteError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func CreateWithdraw(svc gophermart.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		withdrawal, err := bindWithdrawlFromJSON(r)
		if err != nil {
			svc.WriteError(w, err)
			return
		}

		err = svc.PutWithdrawl(r.Context(), *user, withdrawal)
		if err != nil {
			svc.WriteError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetWithdraws(svc gophermart.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		user := auth.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		withdrawals, err := svc.GetWithdrawals(r.Context(), user.ID)
		if err != nil {
			svc.WriteError(w, err)
			return
		}

		if err := responseJSONWithdrawals(w, withdrawals); err != nil {
			svc.WriteError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

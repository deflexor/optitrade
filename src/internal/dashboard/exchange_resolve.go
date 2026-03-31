package dashboard

import (
	"context"
	"fmt"

	"github.com/dfr/optitrade/src/internal/deribit"
)

// cachedExchange avoids rebuilding REST clients on every request when settings are unchanged.
type cachedExchange struct {
	updatedAtMs int64
	x           exchangeReader
}

// exchangeForRequest returns the venue client for the signed-in operator, or (nil, nil) if not configured.
func (s *Server) exchangeForRequest(ctx context.Context) (exchangeReader, error) {
	if s.testXchg != nil {
		return s.testXchg, nil
	}
	user, ok := requestUser(ctx)
	if !ok || s.settings == nil || s.settingsCrypto == nil {
		return nil, nil
	}
	return s.resolveExchange(ctx, user)
}

func (s *Server) resolveExchange(ctx context.Context, username string) (exchangeReader, error) {
	row, err := s.settings.GetDecrypting(ctx, username, s.settingsCrypto)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	if err := validateOperatorSettings(row); err != nil {
		//Incomplete row in DB (e.g. legacy partial write): treat as unavailable.
		return nil, nil
	}

	s.xchgMu.Lock()
	defer s.xchgMu.Unlock()
	if s.xchgCache == nil {
		s.xchgCache = map[string]cachedExchange{}
	}
	if c, ok := s.xchgCache[username]; ok && c.updatedAtMs == row.UpdatedAtMs {
		return c.x, nil
	}
	x, err := s.buildExchange(row)
	if err != nil {
		return nil, err
	}
	s.xchgCache[username] = cachedExchange{updatedAtMs: row.UpdatedAtMs, x: x}
	return x, nil
}

func (s *Server) buildExchange(row *OperatorSettingsRow) (exchangeReader, error) {
	switch row.Provider {
	case ProviderDeribit:
		base := deribit.TestnetRPCBaseURL
		if row.DeribitUseMainnet {
			base = deribit.MainnetRPCBaseURL
		}
		r, err := deribit.NewREST(base, &deribit.Credentials{
			ClientID:     row.Secrets.DeribitClientID,
			ClientSecret: row.Secrets.DeribitClientSecret,
		})
		if err != nil {
			return nil, err
		}
		return r, nil
	case ProviderOKX:
		return newOKXExchange(row)
	default:
		return nil, fmt.Errorf("unsupported provider %q", row.Provider)
	}
}

func (s *Server) invalidateExchangeCache(username string) {
	s.xchgMu.Lock()
	defer s.xchgMu.Unlock()
	if s.xchgCache == nil {
		return
	}
	delete(s.xchgCache, username)
}

// tradingModeForUsername returns test|live from saved operator settings.
func (s *Server) tradingModeForUsername(ctx context.Context, username string) (mode string, ok bool) {
	if s.settings == nil || s.settingsCrypto == nil {
		return "", false
	}
	row, err := s.settings.GetDecrypting(ctx, username, s.settingsCrypto)
	if err != nil || row == nil {
		return "", false
	}
	switch row.Provider {
	case ProviderDeribit:
		if row.DeribitUseMainnet {
			return "live", true
		}
		return "test", true
	case ProviderOKX:
		if row.OKXDemo {
			return "test", true
		}
		return "live", true
	default:
		return "", false
	}
}

package consts

const REQUEST_ID_KEY = "request_id"
const COINCAP_URL = "https://api.coincap.io"
const ASSETS_URI = "/v2/assets"
const RATES_URI = "/v2/rates"
const CANDLES_URI = "/v2/candles"

const RMB_COINCAP_ID = "chinese-yuan-renminbi"

const COINCAP_ASSETS_KEY = "coincap_assets"
const REDIS_NX_EXPIRED_SECONDS = 5
const REDIS_KEY_EXPIRED_SECONDS = 24 * 60 * 60
const HUOBI_BALANCE_KEY = "huobi_balance"

const (
	ANALYZE_ORDER_DEFAULT = iota
	ANALYZE_BEAR_ORDER_FIX_BUY_HOUR
	ANALYZE_BEAR_ORDER_MAX_DEPTH
	ANALYZE_MORE_ORDER_MAX_DEPTH
	ANALYZE_RANDOM_BUY
	ANALYZE_WEEKLY_LOW_PRICE
	ANALYZE_SUNNDAY_RANDOM_BUY
	ANALYZE_CONTRACT_HEDGE
)

const DATA_BASE_DIR = "./data/"

const (
	OrderTypeUnkown = iota
	OrderTypeBuy
	OrderTypeSale
)

const (
	OrderStatusUnkown = iota
	OrderStatusSuccess
	OrderStatusBuy
	OrderStatusCancel
	OrderStatusSettle
)

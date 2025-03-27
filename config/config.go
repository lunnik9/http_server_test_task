package config

type Config struct {
	FetchTimeout  int `env:"FETCH_TIMEOUT" envDefault:"100"`
	MaxURLs       int `env:"MAX_URLS" envDefault:"20"`
	RequestsLimit int `env:"REQUESTS_LIMIT" envDefault:"10"`
	WorkerLimit   int `env:"WORKER_LIMIT" envDefault:"4"`

	Retry          bool `env:"RETRY" envDefault:"true"`
	RetryNum       int  `env:"RETRY_NUM" envDefault:"3"`
	RetryDelay     int  `env:"RETRY_DELAY" envDefault:"30"`
	RetryFillRatio int  `env:"RETRY_FILL_RATIO" envDefault:"80"`

	Cache    bool  `env:"CACHE" envDefault:"true"`
	CacheTTL int64 `env:"CACHE_TTL" envDefault:"30"`
}

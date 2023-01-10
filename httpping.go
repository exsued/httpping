package httpping

import (
    "sync"
    "time"
    "net/http"
)

type context struct {
	stop chan bool
	done chan bool
	err  error
}

func newContext() *context {
	return &context{
		stop: make(chan bool),
		done: make(chan bool),
	}
}

type HttpPinger struct {
	Dests   []string
    Interval float64 
    AlarmInterval float64
    HttpGetTimeout float64
    OnRecv func(int)
    OnFailedRecv func(error)
	OnAlarm func()
	Debug bool

    ctx     *context
	mu      sync.Mutex
	
}

func NewHttpPinger(destinations []string, interval float64, alarminterval float64, httpGetTimeout float64) *HttpPinger {
	return &HttpPinger{
		Dests:   destinations,
        Interval: interval,
        AlarmInterval: alarminterval,
        HttpGetTimeout: httpGetTimeout,
		OnRecv:  nil,
		OnAlarm:  nil,
        OnFailedRecv:  nil,
		Debug:   false,
	}
}

func (p *HttpPinger) Start() error {
    p.mu.Lock()
	p.ctx = newContext()
	p.mu.Unlock()
	p.loop()
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.ctx.err
} 

func (p *HttpPinger) Stop() {
	close(p.ctx.stop)
	<-p.ctx.done
}

func (p *HttpPinger) loop() {
    var err error
    intD := time.Duration(p.Interval) * time.Second
    alrmD:= time.Duration(p.AlarmInterval) * time.Second

    timer := time.AfterFunc(alrmD, func() {
        if p.OnAlarm != nil {
            p.OnAlarm()
        }
    })

    for err == nil {
        for _, dest := range p.Dests { 
            status, err := p.httpPing(dest)
            if err == nil {
                timer.Reset(alrmD)    
                if p.OnRecv != nil {
                    p.OnRecv(status)
                }
            }else if p.OnFailedRecv != nil {
                p.OnFailedRecv(err)
            }
            time.Sleep(intD)  
        }
    }
}

func (p *HttpPinger) httpPing(destination string) (int, error) {
    url := "http://" + destination
    client := http.Client{
        Timeout: time.Duration(p.HttpGetTimeout) * time.Second,
    }
    resp, err := client.Get(url)
    if err != nil {
       return 0, err
    }
    return resp.StatusCode, nil
}

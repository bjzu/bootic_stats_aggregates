package zmq

import(
  "regexp"
  "log"
  data "github.com/bootic/bootic_go_data"
  zmq "github.com/alecthomas/gozmq"
)

type Daemon struct {
  socket zmq.Socket
  observers map[string][]data.EventsChannel
}

func (d *Daemon) listen() {
  for {
    msg, _ := d.socket.Recv(0)
    
    reg, _ := regexp.Compile(`^([^ ]+)?\s+(.+)`)
    
    r := reg.FindStringSubmatch(string(msg))
    
    if len(r) > 1 {
      payload := r[1]
      event, jsonErr := data.Decode([]byte(payload))
      if jsonErr != nil {
        log.Println("Invalid data", jsonErr)
      } else {
       d.Dispatch(event) 
      }
    } else {
      log.Println("Irregular expression", string(msg))
    }
  }
}

func (self *Daemon) SubscribeToType(observer data.EventsChannel, typeStr string) {
  self.observers[typeStr] = append(self.observers[typeStr], observer)
}

func (self *Daemon) Dispatch(event *data.Event) {
  // Dispatch to global observers
  for _, observer := range self.observers["all"] {
    observer <- event
  }
  
  // Dispatch to type observers
  evtStr, _ := event.Get("type").String()
  for _, observer := range self.observers[evtStr] {
    observer <- event
  }
}

func NewZMQSubscriber(host, topic string) (daemon *Daemon, err error) {
  context, _ := zmq.NewContext()
  socket, err := context.NewSocket(zmq.SUB)
  
  socket.SetSockOptString(zmq.SUBSCRIBE, topic)

  socket.Connect(host)
  
  daemon = &Daemon{
    socket: socket,
    observers: make(map[string][]data.EventsChannel),
  }
  
  go daemon.listen()
  
  return
}
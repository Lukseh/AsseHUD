package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
)

//go:embed hud/dist/**
var FrontendFS embed.FS

var (
	modkernel32         = syscall.NewLazyDLL("kernel32.dll")
	procOpenFileMapping = modkernel32.NewProc("OpenFileMappingW")
	procMapViewOfFile   = modkernel32.NewProc("MapViewOfFile")
	procUnmapViewOfFile = modkernel32.NewProc("UnmapViewOfFile")
)

const FILE_MAP_READ = 0x0004

func readSharedMem(name string, size uintptr) ([]byte, error) {
	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}
	handle, _, err := procOpenFileMapping.Call(FILE_MAP_READ, 0, uintptr(unsafe.Pointer(namePtr)))
	if handle == 0 {
		return nil, fmt.Errorf("OpenFileMappingW(%s): %w", name, err)
	}
	defer syscall.CloseHandle(syscall.Handle(handle))

	addr, _, err := procMapViewOfFile.Call(handle, FILE_MAP_READ, 0, 0, size)
	if addr == 0 {
		return nil, fmt.Errorf("MapViewOfFile(%s): %w", name, err)
	}
	defer procUnmapViewOfFile.Call(addr)

	buf := make([]byte, size)
	copy(buf, unsafe.Slice((*byte)(unsafe.Pointer(addr)), size))
	return buf, nil
}

func readInto[T any](name string) (*T, error) {
	var zero T
	size := unsafe.Sizeof(zero)
	buf, err := readSharedMem(name, size)
	if err != nil {
		return nil, err
	}
	return (*T)(unsafe.Pointer(&buf[0])), nil
}

type SPageFilePhysics struct {
	PacketId          int32      `json:"packet_id"`
	Gas               float32    `json:"gas"`
	Brake             float32    `json:"brake"`
	Fuel              float32    `json:"fuel"`
	Gear              int32      `json:"gear"`
	Rpms              int32      `json:"rpm"`
	SteerAngle        float32    `json:"steering"`
	SpeedKmh          float32    `json:"speed"`
	Velocity          [3]float32 `json:"velocity"`
	AccG              [3]float32 `json:"gforce"`
	WheelSlip         [4]float32 `json:"wheel_slip"`
	WheelLoad         [4]float32 `json:"wheel_load"`
	WheelsPressure    [4]float32 `json:"wheel_pressure"`
	WheelAngularSpeed [4]float32 `json:"wheel_speed"`
	TyreWear          [4]float32 `json:"tyre_wear"`
	TyreDirtyLevel    [4]float32 `json:"tyre_dirty"`
	TyreCoreTemp      [4]float32 `json:"tyre_temp"`
	CamberRad         [4]float32 `json:"camber"`
	SuspensionTravel  [4]float32 `json:"suspension_travel"`
	Drs               float32    `json:"drs"`
	TC                float32    `json:"tc"`
	Heading           float32    `json:"heading"`
	Pitch             float32    `json:"pitch"`
	Roll              float32    `json:"roll"`
	CgHeight          float32    `json:"cg_height"`
	CarDamage         [5]float32 `json:"car_damage"`
	NumberOfTyresOut  int32      `json:"tyres_out"`
	PitLimiterOn      int32      `json:"pit_limiter"`
	Abs               float32    `json:"abs"`
}

type SPageFileGraphics struct {
	PacketId           int32      `json:"packet_id"`
	Status             int32      `json:"status"`
	Session            int32      `json:"session"`
	CurrentTime        int32      `json:"current_time_ms"`
	LastTime           int32      `json:"last_time_ms"`
	BestTime           int32      `json:"best_time_ms"`
	SessionTimeLeft    float32    `json:"session_time_left"`
	DistanceTraveled   float32    `json:"distance_traveled"`
	IsInPit            int32      `json:"is_in_pit"`
	CurrentSectorIndex int32      `json:"current_sector"`
	LastSectorTime     int32      `json:"last_sector_time_ms"`
	NumberOfLaps       int32      `json:"laps"`
	TyreCompound       [33]uint16 `json:"tyre_compound"`
	Position           int32      `json:"position"`
	ICurrentTime       int32      `json:"i_current_time_ms"`
	ILastTime          int32      `json:"i_last_time_ms"`
	IBestTime          int32      `json:"i_best_time_ms"`
	FlagColor          int32      `json:"flag"`
	PenaltyTime        float32    `json:"penalty_time"`
	IdealLineOn        int32      `json:"ideal_line"`
	IsInPitLane        int32      `json:"is_in_pit_lane"`
	SurfaceGrip        float32    `json:"surface_grip"`
}

type SPageFileStatic struct {
	SMVersion           [15]uint16 `json:"sm_version"`
	ACVersion           [15]uint16 `json:"ac_version"`
	NumberOfSessions    int32      `json:"sessions"`
	NumCars             int32      `json:"num_cars"`
	CarModel            [33]uint16 `json:"car_model"`
	Track               [33]uint16 `json:"track"`
	PlayerName          [33]uint16 `json:"player_name"`
	PlayerSurname       [33]uint16 `json:"player_surname"`
	PlayerNick          [33]uint16 `json:"player_nick"`
	SectorCount         int32      `json:"sector_count"`
	MaxTorque           float32    `json:"max_torque"`
	MaxPower            float32    `json:"max_power"`
	MaxRpm              int32      `json:"max_rpm"`
	MaxFuel             float32    `json:"max_fuel"`
	SuspensionMaxTravel [4]float32 `json:"suspension_max_travel"`
	TyreRadius          [4]float32 `json:"tyre_radius"`
}

type GameState struct {
	Physics  *SPageFilePhysics
	Graphics *SPageFileGraphics
	Static   *SPageFileStatic
}

type Store struct {
	mu    sync.RWMutex
	state GameState
}

func (s *Store) Update(gs GameState) {
	s.mu.Lock()
	s.state = gs
	s.mu.Unlock()
}

func (s *Store) Get() GameState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func poll(store *Store, interval time.Duration) {
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for range tick.C {
		physics, err := readInto[SPageFilePhysics]("Local\\acpmf_physics")
		if err != nil {
			fmt.Fprintf(os.Stderr, "physics: %v\n", err)
			continue
		}
		graphics, err := readInto[SPageFileGraphics]("Local\\acpmf_graphics")
		if err != nil {
			fmt.Fprintf(os.Stderr, "graphics: %v\n", err)
			continue
		}
		static, err := readInto[SPageFileStatic]("Local\\acpmf_static")
		if err != nil {
			fmt.Fprintf(os.Stderr, "static: %v\n", err)
			continue
		}
		store.Update(GameState{physics, graphics, static})
	}
}

func wsHandler[T any](store *Store, get func(GameState) *T, interval time.Duration) func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		tick := time.NewTicker(interval)
		defer tick.Stop()
		for range tick.C {
			data := get(store.Get())
			if data == nil {
				continue
			}
			if err := c.WriteJSON(data); err != nil {
				log.Println("ws write:", err)
				return
			}
		}
	}
}

type ServerConfig struct {
	Port            int    `json:"port"`
	Host            string `json:"host"`
	LogLevel        string `json:"loglevel"`
	PollingInterval int    `json:"polling_interval_ms"`
}
type FrontendConfig struct {
	DefaultIcon   bool `json:"default_icon"`
	DefaultText   bool `json:"default_text"`
	WheelAngle    int  `json:"wheel_max_angle"`
	GraphDuration int  `json:"graph_duration"`
}
type Config struct {
	Server   ServerConfig   `json:"server"`
	Frontend FrontendConfig `json:"page"`
}

var (
	cfgOnly bool
)

func main() {
	flag.BoolVar(&cfgOnly, "C", false, "Generates config.json and exits.")
	flag.Parse()

	f, err := os.OpenFile("config.json", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	rawF, err := io.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	// Default one
	cfg := Config{
		Server: ServerConfig{
			LogLevel:        "none",
			Port:            16440,
			Host:            "localhost",
			PollingInterval: 30,
		},
		Frontend: FrontendConfig{
			DefaultIcon:   true,
			DefaultText:   true,
			WheelAngle:    900,
			GraphDuration: 5,
		},
	}
	if len(rawF) == 0 {
		js, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			log.Fatalln(err)
		}
		if err := f.Truncate(0); err != nil {
			log.Fatalln(err)
		}
		if _, err := f.Seek(0, 0); err != nil {
			log.Fatalln(err)
		}
		if _, err := f.Write(js); err != nil {
			log.Fatalln(err)
		}
		if err := f.Sync(); err != nil {
			log.Fatalln(err)
		}
	} else {
		if err := json.Unmarshal(rawF, &cfg); err != nil {
			log.Fatalln(err)
		}
	}

	if cfgOnly {
		log.Println("Saved config.json")
		os.Exit(0)
	}

	store := &Store{}
	go poll(store, time.Duration(cfg.Server.PollingInterval)*time.Millisecond)
	app := fiber.New()
	app.Use(cors.New(cors.Config{AllowOrigins: []string{"*"}}))
	app.Use(logger.New(logger.Config{
		Skip: func(c fiber.Ctx) bool {
			if cfg.Server.LogLevel == "none" {
				return true
			}
			return false
		},
	}))
	app.Use("/ws", func(c fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/config.json", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"page": cfg.Frontend, "polling_rate": cfg.Server.PollingInterval})
	})
	ws := app.Group("/ws")
	ws.Get("/physics", websocket.New(wsHandler(store, func(gs GameState) *SPageFilePhysics { return gs.Physics }, time.Duration(cfg.Server.PollingInterval)*time.Millisecond)))
	ws.Get("/graphics", websocket.New(wsHandler(store, func(gs GameState) *SPageFileGraphics { return gs.Graphics }, time.Duration(cfg.Server.PollingInterval)*time.Millisecond)))
	ws.Get("/static", websocket.New(wsHandler(store, func(gs GameState) *SPageFileStatic { return gs.Static }, time.Duration(cfg.Server.PollingInterval)*time.Millisecond)))
	frontendSub, err := fs.Sub(FrontendFS, "hud/dist")
	if err != nil {
		log.Fatalln(err)
	}
	app.Use(static.New("", static.Config{
		FS:         frontendSub,
		IndexNames: []string{"index.html"},
		Browse:     true,
	}))
	c, _ := json.MarshalIndent(cfg, "", "  ")
	log.Println(string(c))
	log.Printf("Your app is working on http://%s:%d/\n", cfg.Server.Host, cfg.Server.Port)
	log.Fatalln(app.Listen(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port), fiber.ListenConfig{DisableStartupMessage: true}))
}

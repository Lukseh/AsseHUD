export interface AppConfig {
  default_icon: boolean
  default_text: boolean
  wheel_max_angle: number
}
export interface StaticData {
  sm_version: number[]
  ac_version: number[]
  sessions: number
  num_cars: number
  car_model: number[]
  track: number[]
  player_name: number[]
  player_surname: number[]
  player_nick: number[]
  sector_count: number
  max_torque: number
  max_power: number
  max_rpm: number
  max_fuel: number
  suspension_max_travel: number[]
  tyre_radius: number[]
}
export interface PhysicsData {
  packet_id: number;
  gas: number;
  brake: number;
  fuel: number;
  gear: number;
  rpm: number;
  steering: number;
  speed: number;
  velocity: number[];
  gforce: number[];
  wheel_slip: number[];
  wheel_load: number[];
  wheel_pressure: number[];
  wheel_speed: number[];
  tyre_wear: number[];
  tyre_dirty: number[];
  tyre_temp: number[];
  camber: number[];
  suspension_travel: number[];
  drs: number;
  tc: number;
  heading: number;
  pitch: number;
  roll: number;
  cg_height: number;
  car_damage: number[];
  tyres_out: number;
  pit_limiter: number;
  abs: number;
}

export interface Video {
  id: string;
  url: string;
  title: string;
  duration: number;
  file_path: string;
  thumbnail_url: string;
  status:
    | "pending"
    | "downloading"
    | "transcribing"
    | "analyzing"
    | "completed"
    | "error";
  created_at: string;
  updated_at: string;
}

export interface Segment {
  start: number;
  end: number;
  text: string;
}

export interface Transcript {
  id: string;
  video_id: string;
  language: string;
  segments: Segment[];
  full_text: string;
  created_at: string;
}

export interface SuggestedClip {
  id: string;
  video_id: string;
  start_time: number;
  end_time: number;
  title: string;
  description: string;
  score: number;
  reason: string;
}

export interface SubtitleConfig {
  text: string;
  start_time: number;
  end_time: number;
  font_family: string;
  font_size: number;
  font_weight: number;
  color: string;
  bg_color: string;
  bg_opacity: number;
  position: "top" | "center" | "bottom";
  bold: boolean;
  italic: boolean;
  border_radius: number;
  shadow_blur: number;
  transition:
    | "none"
    | "pop"
    | "fade"
    | "slide"
    | "bounce"
    | "zoom"
    | "blur"
    | "scale"
    | "rotate"
    | "flip"
    | "elastic"
    | "spring";
  active_text_color: string;
  // Campos opcionales para la sincronización con la transcripción
  segmentId?: number;
  originalStart?: number;
  originalEnd?: number;
  // Para efecto karaoke palabra por palabra
  wordIndex?: number;
  totalWords?: number;
  // Para grupos de palabras
  groupIndex?: number;
  totalGroups?: number;
}

export interface Clip {
  id: string;
  video_id: string;
  title: string;
  start_time: number;
  end_time: number;
  file_path: string;
  status: "processing" | "completed" | "error";
  subtitles: SubtitleConfig[];
  created_at: string;
  completed_at?: string;
}

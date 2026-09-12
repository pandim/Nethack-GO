package game

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

// =============================================================================
// ЛИНЕЙНАЯ ГРОМКОСТЬ
// =============================================================================
type linearVolume struct {
	streamer beep.Streamer
	volume   float64 // 0.0 (тишина) до 1.0 (полная громкость)
}

type musicPlaylist struct {
	tracks     []string
	current    beep.StreamSeekCloser
	sampleRate beep.SampleRate
	err        error
	closed     bool
}

func (p *musicPlaylist) Stream(samples [][2]float64) (n int, ok bool) {
	for len(samples) > 0 && !p.closed {
		if p.current == nil && !p.openRandomTrack() {
			return n, false
		}

		streamed, playing := p.current.Stream(samples)
		n += streamed
		samples = samples[streamed:]
		if playing {
			return n, true
		}

		p.current.Close()
		p.current = nil
	}

	return n, !p.closed && n > 0
}

func (p *musicPlaylist) Err() error {
	if p.err != nil {
		return p.err
	}
	if p.current != nil {
		return p.current.Err()
	}
	return nil
}

func (p *musicPlaylist) Close() error {
	p.closed = true
	if p.current != nil {
		err := p.current.Close()
		p.current = nil
		return err
	}
	return nil
}

func (p *musicPlaylist) skipTrack() {
	if p.current != nil {
		p.current.Close()
		p.current = nil
	}
	p.err = nil
}

func (p *musicPlaylist) openRandomTrack() bool {
	for range p.tracks {
		path := p.tracks[rand.IntN(len(p.tracks))]
		file, err := os.Open(path)
		if err != nil {
			p.err = err
			continue
		}

		streamer, format, err := mp3.Decode(file)
		if err != nil {
			file.Close()
			p.err = err
			continue
		}

		p.current = streamer
		p.sampleRate = format.SampleRate
		p.err = nil
		return true
	}
	return false
}

func findMusicTracks() ([]string, error) {
	for _, directory := range []string{"Music", "./Music", "../Music"} {
		entries, err := os.ReadDir(directory)
		if err != nil {
			continue
		}

		tracks := make([]string, 0, len(entries))
		for _, entry := range entries {
			if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".mp3") {
				tracks = append(tracks, filepath.Join(directory, entry.Name()))
			}
		}
		if len(tracks) > 0 {
			return tracks, nil
		}
	}

	return nil, fmt.Errorf("в папке Music не найдено mp3-файлов")
}

func (v *linearVolume) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = v.streamer.Stream(samples)
	for i := range samples[:n] {
		samples[i][0] *= v.volume
		samples[i][1] *= v.volume
	}
	return n, ok
}

func (v *linearVolume) Err() error {
	return v.streamer.Err()
}

// =============================================================================
// УПРАВЛЕНИЕ МУЗЫКОЙ
// =============================================================================

// initMusic — инициализирует случайный плейлист фоновой музыки.
func (g *Game) initMusic() {
	// Останавливаем старую музыку (если играла)
	g.stopMusic()

	if !g.musicEnabled {
		g.logAndSync("MUSIC: Отключено")
		return
	}

	tracks, err := findMusicTracks()
	if err != nil {
		g.logAndSync("MUSIC: %v", err)
		g.musicEnabled = false
		return
	}

	playlist := &musicPlaylist{tracks: tracks}
	if !playlist.openRandomTrack() {
		g.logAndSync("MUSIC: Не удалось открыть ни один mp3-файл: %v", playlist.Err())
		g.musicEnabled = false
		return
	}
	// Инициализируем звуковое устройство
	err = speaker.Init(playlist.sampleRate, playlist.sampleRate.N(time.Second/10))
	if err != nil {
		playlist.Close()
		g.logAndSync("MUSIC: Ошибка инициализации speaker: %v", err)
		g.musicEnabled = false
		return
	}

	// Оборачиваем в линейную обёртку громкости
	g.musicVolume = &linearVolume{
		streamer: playlist,
		volume:   g.musicLevel,
	}

	// Оборачиваем в beep.Ctrl для паузы/возобновления
	g.musicCtrl = &beep.Ctrl{
		Streamer: g.musicVolume,
		Paused:   false,
	}

	// Запускаем воспроизведение
	speaker.Play(g.musicCtrl)

	// Сохраняем ссылки для последующего закрытия
	g.musicStreamer = playlist

	g.logAndSync("MUSIC: Запущена (rate=%d, volume=%.2f, tracks=%d)", playlist.sampleRate, g.musicLevel, len(tracks))
}

// stopMusic — полностью останавливает музыку и освобождает ресурсы.
func (g *Game) stopMusic() {
	speaker.Clear()

	if g.musicStreamer != nil {
		g.musicStreamer.Close()
		g.musicStreamer = nil
	}

	g.musicVolume = nil
	g.musicCtrl = nil

	speaker.Close()
	g.logAndSync("MUSIC: Остановлена")
}

// toggleMusic — включает/выключает музыку.
func (g *Game) toggleMusic() {
	if !g.musicEnabled || g.musicCtrl == nil {
		g.addMessage("Музыка недоступна.")
		return
	}

	speaker.Lock()
	g.musicCtrl.Paused = !g.musicCtrl.Paused
	paused := g.musicCtrl.Paused
	speaker.Unlock()

	if paused {
		g.addMessage("Музыка выключена.")
	} else {
		g.addMessage("Музыка включена.")
	}
}

func (g *Game) nextMusicTrack() {
	if !g.musicEnabled || g.musicCtrl == nil {
		g.addMessage("Музыка недоступна.")
		return
	}

	playlist, ok := g.musicStreamer.(*musicPlaylist)
	if !ok {
		g.addMessage("Музыка недоступна.")
		return
	}

	speaker.Lock()
	playlist.skipTrack()
	speaker.Unlock()
	g.addMessage("Включен следующий трек.")
}

// changeMusicVolume — меняет громкость музыки.
func (g *Game) changeMusicVolume(delta float64) {
	if !g.musicEnabled || g.musicVolume == nil {
		g.addMessage("Музыка недоступна.")
		return
	}

	g.musicLevel += delta
	if g.musicLevel < 0 {
		g.musicLevel = 0
	}
	if g.musicLevel > 1.0 {
		g.musicLevel = 1.0
	}

	speaker.Lock()
	g.musicVolume.volume = g.musicLevel
	speaker.Unlock()

	percent := int(g.musicLevel * 100)
	g.addMessage(fmt.Sprintf("Громкость: %d%%", percent))
}

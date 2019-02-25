package main

import (
	"fmt"
	"github.com/go-audio/wav"
	"io"
	"os"
	"path"
	"time"
)

func (s server) saveWAV(r io.ReadSeeker, topic, user string) error {
	d := wav.NewDecoder(r)
	buff, err := d.FullPCMBuffer()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s-%s-%d.wav", user, topic, time.Now().Unix())

	f, err := os.OpenFile(
		path.Join("./audio/", filename),
		os.O_WRONLY|os.O_CREATE,
		0664)
	if err != nil {
		return err
	}
	defer f.Close()

	e := wav.NewEncoder(
		f,
		buff.Format.SampleRate,
		int(d.BitDepth),
		buff.Format.NumChannels,
		int(d.WavAudioFormat))
	if err := e.Write(buff); err != nil {
		return err
	}
	if err := e.Close(); err != nil {
		return err
	}

	return s.updateTopic(user, Topic{
		User:     user,
		Topic:    topic,
		Filename: filename,
	})
}

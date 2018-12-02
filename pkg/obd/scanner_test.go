package obd

import (
	"testing"
	"time"

	"github.com/rzetterberg/elmobd"
	"github.com/stretchr/testify/require"
)

func getCommands(dev *elmobd.Device) ([]elmobd.OBDCommand, error) {
	supported, err := dev.CheckSupportedCommands()
	if err != nil {
		return nil, err
	}
	allCommands := elmobd.GetSensorCommands()
	commands := supported.FilterSupported(allCommands)
	return commands, nil
}
func TestScanner(t *testing.T) {
	dev, _ := elmobd.NewTestDevice("", false)
	cmds, _ := getCommands(dev)
	l := len(cmds)

	s := NewScanner(dev, cmds, time.Millisecond*100, 100)

	go func() {
		timer := time.NewTimer(time.Second)
		<-timer.C
		s.Stop()
	}()

	time.Sleep(time.Millisecond * 200)

	c := s.C()
	require.NotEmpty(t, c)
	for m := range c {
		require.Len(t, m, l)
	}
	require.Empty(t, c)
}

func TestScannerMeasurement(t *testing.T) {
	startTime := time.Now()
	dev, _ := elmobd.NewTestDevice("", false)
	cmd := elmobd.NewEngineRPM()
	s := NewScanner(dev, []elmobd.OBDCommand{cmd}, time.Second, 0)

	m, ok := <-s.C()
	s.Stop()

	require.True(t, ok)
	require.Len(t, m, 1)
	require.Equal(t, cmd.ModeID(), m[0].ModeID)
	require.Equal(t, byte(cmd.ParameterID()), m[0].ParameterID)
	require.Equal(t, cmd.Key(), m[0].Key)
	require.Equal(t, float64(cmd.Value), m[0].Value)
	require.True(t, startTime.Nanosecond() < m[0].Time.Nanosecond())

}

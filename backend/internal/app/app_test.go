package app

import "testing"

func TestAppBootstrap_NewComDependencias_ResultadoEsperado(t *testing.T) {
	instance := New("0", t.TempDir())
	defer instance.Analytics.Shutdown()

	if instance.Server == nil {
		t.Fatal("expected non-nil server")
	}
	if instance.Server.Handler == nil {
		t.Fatal("expected non-nil server handler")
	}
}

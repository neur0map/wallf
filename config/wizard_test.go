package config

import "testing"

func TestWizardInitialState(t *testing.T) {
	w := NewWizard()
	if w.step != wizardStepDir {
		t.Errorf("expected wizardStepDir (0), got %v", w.step)
	}
}

func TestWizardDefaults(t *testing.T) {
	w := NewWizard()
	cfg := w.Config()
	if cfg.General.DownloadDir == "" {
		t.Error("expected non-empty DownloadDir")
	}
	if cfg.General.MinResolution == "" {
		t.Error("expected non-empty MinResolution")
	}
}

func TestWizardStepCount(t *testing.T) {
	if wizardStepConfirm != 2 {
		t.Errorf("expected wizardStepConfirm == 2, got %d", wizardStepConfirm)
	}
}

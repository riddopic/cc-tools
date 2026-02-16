package config

// Values represents the concrete configuration structure.
type Values struct {
	Validate       ValidateValues       `json:"validate"`
	Notifications  NotificationsValues  `json:"notifications"`
	Compact        CompactValues        `json:"compact"`
	Notify         NotifyValues         `json:"notify"`
	Observe        ObserveValues        `json:"observe"`
	Learning       LearningValues       `json:"learning"`
	PreCommit      PreCommitValues      `json:"pre_commit_reminder"`
	PackageManager PackageManagerValues `json:"package_manager"`
}

// NotificationsValues represents notification-related settings.
type NotificationsValues struct {
	NtfyTopic string `json:"ntfy_topic"`
}

// ValidateValues represents validate-related settings.
type ValidateValues struct {
	Timeout  int `json:"timeout"`
	Cooldown int `json:"cooldown"`
}

// CompactValues represents compact context reminder settings.
type CompactValues struct {
	Threshold        int `json:"threshold"`
	ReminderInterval int `json:"reminder_interval"`
}

// NotifyValues represents notification dispatch settings.
type NotifyValues struct {
	QuietHours QuietHoursValues `json:"quiet_hours"`
	Audio      AudioValues      `json:"audio"`
	Desktop    DesktopValues    `json:"desktop"`
}

// QuietHoursValues represents quiet hours configuration.
type QuietHoursValues struct {
	Enabled bool   `json:"enabled"`
	Start   string `json:"start"`
	End     string `json:"end"`
}

// AudioValues represents audio notification settings.
type AudioValues struct {
	Enabled   bool   `json:"enabled"`
	Directory string `json:"directory"`
}

// DesktopValues represents desktop notification settings.
type DesktopValues struct {
	Enabled bool `json:"enabled"`
}

// ObserveValues represents file observation settings.
type ObserveValues struct {
	Enabled       bool `json:"enabled"`
	MaxFileSizeMB int  `json:"max_file_size_mb"`
}

// LearningValues represents learning extraction settings.
type LearningValues struct {
	MinSessionLength  int    `json:"min_session_length"`
	LearnedSkillsPath string `json:"learned_skills_path"`
}

// PreCommitValues represents pre-commit reminder settings.
type PreCommitValues struct {
	Enabled bool   `json:"enabled"`
	Command string `json:"command"`
}

// PackageManagerValues represents package manager preference settings.
type PackageManagerValues struct {
	Preferred string `json:"preferred"`
}

// convertValidateFromMap extracts validate settings from a map config.
func convertValidateFromMap(v *ValidateValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["validate"].(map[string]any)
	if !sectionOk {
		return
	}
	if timeout, timeoutOk := section["timeout"].(float64); timeoutOk {
		v.Timeout = int(timeout)
	}
	if cooldown, cooldownOk := section["cooldown"].(float64); cooldownOk {
		v.Cooldown = int(cooldown)
	}
}

// convertNotificationsFromMap extracts notification settings from a map config.
func convertNotificationsFromMap(n *NotificationsValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["notifications"].(map[string]any)
	if !sectionOk {
		return
	}
	if topic, topicOk := section["ntfy_topic"].(string); topicOk {
		n.NtfyTopic = topic
	}
}

// convertCompactFromMap extracts compact settings from a map config.
func convertCompactFromMap(c *CompactValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["compact"].(map[string]any)
	if !sectionOk {
		return
	}
	if threshold, thresholdOk := section["threshold"].(float64); thresholdOk {
		c.Threshold = int(threshold)
	}
	if interval, intervalOk := section["reminder_interval"].(float64); intervalOk {
		c.ReminderInterval = int(interval)
	}
}

// convertNotifyFromMap extracts notify settings (quiet hours, audio, desktop) from a map.
func convertNotifyFromMap(n *NotifyValues, notifyMap map[string]any) {
	if qhMap, qhOk := notifyMap["quiet_hours"].(map[string]any); qhOk {
		if enabled, enabledOk := qhMap["enabled"].(bool); enabledOk {
			n.QuietHours.Enabled = enabled
		}
		if start, startOk := qhMap["start"].(string); startOk {
			n.QuietHours.Start = start
		}
		if end, endOk := qhMap["end"].(string); endOk {
			n.QuietHours.End = end
		}
	}
	if audioMap, audioOk := notifyMap["audio"].(map[string]any); audioOk {
		if enabled, enabledOk := audioMap["enabled"].(bool); enabledOk {
			n.Audio.Enabled = enabled
		}
		if dir, dirOk := audioMap["directory"].(string); dirOk {
			n.Audio.Directory = dir
		}
	}
	if desktopMap, desktopOk := notifyMap["desktop"].(map[string]any); desktopOk {
		if enabled, enabledOk := desktopMap["enabled"].(bool); enabledOk {
			n.Desktop.Enabled = enabled
		}
	}
}

// convertObserveFromMap extracts observe settings from a map config.
func convertObserveFromMap(o *ObserveValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["observe"].(map[string]any)
	if !sectionOk {
		return
	}
	if enabled, enabledOk := section["enabled"].(bool); enabledOk {
		o.Enabled = enabled
	}
	if maxSize, maxSizeOk := section["max_file_size_mb"].(float64); maxSizeOk {
		o.MaxFileSizeMB = int(maxSize)
	}
}

// convertLearningFromMap extracts learning settings from a map config.
func convertLearningFromMap(l *LearningValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["learning"].(map[string]any)
	if !sectionOk {
		return
	}
	if minLen, minLenOk := section["min_session_length"].(float64); minLenOk {
		l.MinSessionLength = int(minLen)
	}
	if path, pathOk := section["learned_skills_path"].(string); pathOk {
		l.LearnedSkillsPath = path
	}
}

// convertPreCommitFromMap extracts pre-commit reminder settings from a map config.
func convertPreCommitFromMap(p *PreCommitValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["pre_commit_reminder"].(map[string]any)
	if !sectionOk {
		return
	}
	if enabled, enabledOk := section["enabled"].(bool); enabledOk {
		p.Enabled = enabled
	}
	if cmd, cmdOk := section["command"].(string); cmdOk {
		p.Command = cmd
	}
}

// convertPackageManagerFromMap extracts package manager settings from a map config.
func convertPackageManagerFromMap(pm *PackageManagerValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["package_manager"].(map[string]any)
	if !sectionOk {
		return
	}
	if preferred, preferredOk := section["preferred"].(string); preferredOk {
		pm.Preferred = preferred
	}
}

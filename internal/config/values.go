package config

import "strconv"

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
	Drift          DriftValues          `json:"drift"`
	StopReminder   StopReminderValues   `json:"stop_reminder"`
	Instinct       InstinctValues       `json:"instinct"`
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

// DriftValues represents drift detection settings.
type DriftValues struct {
	Enabled   bool    `json:"enabled"`
	MinEdits  int     `json:"min_edits"`
	Threshold float64 `json:"threshold"`
}

// StopReminderValues represents stop event reminder settings.
type StopReminderValues struct {
	Enabled  bool `json:"enabled"`
	Interval int  `json:"interval"`
	WarnAt   int  `json:"warn_at"`
}

// InstinctValues represents instinct management settings.
type InstinctValues struct {
	PersonalPath     string  `json:"personal_path"`
	InheritedPath    string  `json:"inherited_path"`
	MinConfidence    float64 `json:"min_confidence"`
	AutoApprove      float64 `json:"auto_approve"`
	DecayRate        float64 `json:"decay_rate"`
	MaxInstincts     int     `json:"max_instincts"`
	ClusterThreshold int     `json:"cluster_threshold"`
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

// getExtendedValue returns a drift or stop_reminder value as a string.
func (v *Values) getExtendedValue(key string) (string, bool, error) {
	switch key {
	case keyDriftEnabled:
		return strconv.FormatBool(v.Drift.Enabled), true, nil
	case keyDriftMinEdits:
		return strconv.Itoa(v.Drift.MinEdits), true, nil
	case keyDriftThreshold:
		return strconv.FormatFloat(v.Drift.Threshold, 'f', -1, 64), true, nil
	case keyStopReminderEnabled:
		return strconv.FormatBool(v.StopReminder.Enabled), true, nil
	case keyStopReminderInterval:
		return strconv.Itoa(v.StopReminder.Interval), true, nil
	case keyStopReminderWarnAt:
		return strconv.Itoa(v.StopReminder.WarnAt), true, nil
	case keyInstinctPersonalPath:
		return v.Instinct.PersonalPath, true, nil
	case keyInstinctInheritedPath:
		return v.Instinct.InheritedPath, true, nil
	case keyInstinctMinConfidence:
		return strconv.FormatFloat(v.Instinct.MinConfidence, 'f', -1, 64), true, nil
	case keyInstinctAutoApprove:
		return strconv.FormatFloat(v.Instinct.AutoApprove, 'f', -1, 64), true, nil
	case keyInstinctDecayRate:
		return strconv.FormatFloat(v.Instinct.DecayRate, 'f', -1, 64), true, nil
	case keyInstinctMaxInstincts:
		return strconv.Itoa(v.Instinct.MaxInstincts), true, nil
	case keyInstinctClusterThreshold:
		return strconv.Itoa(v.Instinct.ClusterThreshold), true, nil
	default:
		return "", false, nil
	}
}

// setExtendedField sets a drift or stop_reminder field from a string value.
func (v *Values) setExtendedField(key, value string) (bool, error) {
	switch key {
	case keyDriftEnabled:
		return true, setBoolField(&v.Drift.Enabled, value)
	case keyDriftMinEdits:
		return true, setIntField(&v.Drift.MinEdits, value)
	case keyDriftThreshold:
		return true, setFloatField(&v.Drift.Threshold, value)
	case keyStopReminderEnabled:
		return true, setBoolField(&v.StopReminder.Enabled, value)
	case keyStopReminderInterval:
		return true, setIntField(&v.StopReminder.Interval, value)
	case keyStopReminderWarnAt:
		return true, setIntField(&v.StopReminder.WarnAt, value)
	case keyInstinctPersonalPath:
		v.Instinct.PersonalPath = value
		return true, nil
	case keyInstinctInheritedPath:
		v.Instinct.InheritedPath = value
		return true, nil
	case keyInstinctMinConfidence:
		return true, setFloatField(&v.Instinct.MinConfidence, value)
	case keyInstinctAutoApprove:
		return true, setFloatField(&v.Instinct.AutoApprove, value)
	case keyInstinctDecayRate:
		return true, setFloatField(&v.Instinct.DecayRate, value)
	case keyInstinctMaxInstincts:
		return true, setIntField(&v.Instinct.MaxInstincts, value)
	case keyInstinctClusterThreshold:
		return true, setIntField(&v.Instinct.ClusterThreshold, value)
	default:
		return false, nil
	}
}

// resetExtended resets drift or stop_reminder fields to their defaults.
func (v *Values) resetExtended(key string, defaults *Values) bool {
	switch key {
	case keyDriftEnabled:
		v.Drift.Enabled = defaults.Drift.Enabled
	case keyDriftMinEdits:
		v.Drift.MinEdits = defaults.Drift.MinEdits
	case keyDriftThreshold:
		v.Drift.Threshold = defaults.Drift.Threshold
	case keyStopReminderEnabled:
		v.StopReminder.Enabled = defaults.StopReminder.Enabled
	case keyStopReminderInterval:
		v.StopReminder.Interval = defaults.StopReminder.Interval
	case keyStopReminderWarnAt:
		v.StopReminder.WarnAt = defaults.StopReminder.WarnAt
	case keyInstinctPersonalPath:
		v.Instinct.PersonalPath = defaults.Instinct.PersonalPath
	case keyInstinctInheritedPath:
		v.Instinct.InheritedPath = defaults.Instinct.InheritedPath
	case keyInstinctMinConfidence:
		v.Instinct.MinConfidence = defaults.Instinct.MinConfidence
	case keyInstinctAutoApprove:
		v.Instinct.AutoApprove = defaults.Instinct.AutoApprove
	case keyInstinctDecayRate:
		v.Instinct.DecayRate = defaults.Instinct.DecayRate
	case keyInstinctMaxInstincts:
		v.Instinct.MaxInstincts = defaults.Instinct.MaxInstincts
	case keyInstinctClusterThreshold:
		v.Instinct.ClusterThreshold = defaults.Instinct.ClusterThreshold
	default:
		return false
	}
	return true
}

// convertDriftFromMap extracts drift detection settings from a map config.
func convertDriftFromMap(d *DriftValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["drift"].(map[string]any)
	if !sectionOk {
		return
	}
	if enabled, enabledOk := section["enabled"].(bool); enabledOk {
		d.Enabled = enabled
	}
	if minEdits, minEditsOk := section["min_edits"].(float64); minEditsOk {
		d.MinEdits = int(minEdits)
	}
	if threshold, thresholdOk := section["threshold"].(float64); thresholdOk {
		d.Threshold = threshold
	}
}

// convertStopReminderFromMap extracts stop reminder settings from a map config.
func convertStopReminderFromMap(sr *StopReminderValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["stop_reminder"].(map[string]any)
	if !sectionOk {
		return
	}
	if enabled, enabledOk := section["enabled"].(bool); enabledOk {
		sr.Enabled = enabled
	}
	if interval, intervalOk := section["interval"].(float64); intervalOk {
		sr.Interval = int(interval)
	}
	if warnAt, warnAtOk := section["warn_at"].(float64); warnAtOk {
		sr.WarnAt = int(warnAt)
	}
}

// convertInstinctFromMap extracts instinct settings from a map config.
func convertInstinctFromMap(i *InstinctValues, mapConfig map[string]any) {
	section, sectionOk := mapConfig["instinct"].(map[string]any)
	if !sectionOk {
		return
	}
	if personalPath, ok := section["personal_path"].(string); ok {
		i.PersonalPath = personalPath
	}
	if inheritedPath, ok := section["inherited_path"].(string); ok {
		i.InheritedPath = inheritedPath
	}
	if minConf, ok := section["min_confidence"].(float64); ok {
		i.MinConfidence = minConf
	}
	if autoApprove, ok := section["auto_approve"].(float64); ok {
		i.AutoApprove = autoApprove
	}
	if decayRate, ok := section["decay_rate"].(float64); ok {
		i.DecayRate = decayRate
	}
	if maxInstincts, ok := section["max_instincts"].(float64); ok {
		i.MaxInstincts = int(maxInstincts)
	}
	if clusterThreshold, ok := section["cluster_threshold"].(float64); ok {
		i.ClusterThreshold = int(clusterThreshold)
	}
}

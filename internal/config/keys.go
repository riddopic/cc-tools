package config

import "strconv"

// Configuration keys.
const (
	keyValidateTimeout        = "validate.timeout"
	keyValidateCooldown       = "validate.cooldown"
	keyNotificationsNtfyTopic = "notifications.ntfy_topic"

	keyCompactThreshold        = "compact.threshold"
	keyCompactReminderInterval = "compact.reminder_interval"

	keyNotifyQuietHoursEnabled = "notify.quiet_hours.enabled"
	keyNotifyQuietHoursStart   = "notify.quiet_hours.start"
	keyNotifyQuietHoursEnd     = "notify.quiet_hours.end"
	keyNotifyAudioEnabled      = "notify.audio.enabled"
	keyNotifyAudioDirectory    = "notify.audio.directory"
	keyNotifyDesktopEnabled    = "notify.desktop.enabled"

	keyObserveEnabled       = "observe.enabled"
	keyObserveMaxFileSizeMB = "observe.max_file_size_mb"

	keyLearningMinSessionLength  = "learning.min_session_length"
	keyLearningLearnedSkillsPath = "learning.learned_skills_path"

	keyPreCommitEnabled = "pre_commit_reminder.enabled"
	keyPreCommitCommand = "pre_commit_reminder.command"

	keyPackageManagerPreferred = "package_manager.preferred"

	keyDriftEnabled   = "drift.enabled"
	keyDriftMinEdits  = "drift.min_edits"
	keyDriftThreshold = "drift.threshold"

	keyStopReminderEnabled  = "stop_reminder.enabled"
	keyStopReminderInterval = "stop_reminder.interval"
	keyStopReminderWarnAt   = "stop_reminder.warn_at"

	keyInstinctPersonalPath     = "instinct.personal_path"
	keyInstinctInheritedPath    = "instinct.inherited_path"
	keyInstinctMinConfidence    = "instinct.min_confidence"
	keyInstinctAutoApprove      = "instinct.auto_approve"
	keyInstinctDecayRate        = "instinct.decay_rate"
	keyInstinctMaxInstincts     = "instinct.max_instincts"
	keyInstinctClusterThreshold = "instinct.cluster_threshold"
)

const (
	defaultValidateTimeout  = 60
	defaultValidateCooldown = 5

	defaultCompactThreshold        = 50
	defaultCompactReminderInterval = 25

	defaultNotifyQuietHoursEnabled = true
	defaultNotifyQuietHoursStart   = "21:00"
	defaultNotifyQuietHoursEnd     = "07:30"
	defaultNotifyAudioEnabled      = true
	defaultNotifyAudioDirectory    = "~/.claude/audio"
	defaultNotifyDesktopEnabled    = true

	defaultObserveEnabled       = true
	defaultObserveMaxFileSizeMB = 10

	defaultLearningMinSessionLength  = 10
	defaultLearningLearnedSkillsPath = ".claude/skills/learned"

	defaultPreCommitEnabled = true
	defaultPreCommitCommand = "task pre-commit"

	defaultPackageManagerPreferred = ""

	defaultDriftEnabled   = true
	defaultDriftMinEdits  = 6
	defaultDriftThreshold = 0.2

	defaultStopReminderEnabled  = true
	defaultStopReminderInterval = 20
	defaultStopReminderWarnAt   = 50

	defaultInstinctPersonalPath     = "~/.config/cc-tools/instincts/personal"
	defaultInstinctInheritedPath    = "~/.config/cc-tools/instincts/inherited"
	defaultInstinctMinConfidence    = 0.3
	defaultInstinctAutoApprove      = 0.7
	defaultInstinctDecayRate        = 0.02
	defaultInstinctMaxInstincts     = 100
	defaultInstinctClusterThreshold = 3
)

// GetDefaultConfig returns the default configuration values.
func GetDefaultConfig() *Values {
	return &Values{
		Validate: ValidateValues{
			Timeout:  defaultValidateTimeout,
			Cooldown: defaultValidateCooldown,
		},
		Notifications: NotificationsValues{
			NtfyTopic: "",
		},
		Compact: CompactValues{
			Threshold:        defaultCompactThreshold,
			ReminderInterval: defaultCompactReminderInterval,
		},
		Notify: NotifyValues{
			QuietHours: QuietHoursValues{
				Enabled: defaultNotifyQuietHoursEnabled,
				Start:   defaultNotifyQuietHoursStart,
				End:     defaultNotifyQuietHoursEnd,
			},
			Audio: AudioValues{
				Enabled:   defaultNotifyAudioEnabled,
				Directory: defaultNotifyAudioDirectory,
			},
			Desktop: DesktopValues{
				Enabled: defaultNotifyDesktopEnabled,
			},
		},
		Observe: ObserveValues{
			Enabled:       defaultObserveEnabled,
			MaxFileSizeMB: defaultObserveMaxFileSizeMB,
		},
		Learning: LearningValues{
			MinSessionLength:  defaultLearningMinSessionLength,
			LearnedSkillsPath: defaultLearningLearnedSkillsPath,
		},
		PreCommit: PreCommitValues{
			Enabled: defaultPreCommitEnabled,
			Command: defaultPreCommitCommand,
		},
		PackageManager: PackageManagerValues{
			Preferred: defaultPackageManagerPreferred,
		},
		Drift: DriftValues{
			Enabled:   defaultDriftEnabled,
			MinEdits:  defaultDriftMinEdits,
			Threshold: defaultDriftThreshold,
		},
		StopReminder: StopReminderValues{
			Enabled:  defaultStopReminderEnabled,
			Interval: defaultStopReminderInterval,
			WarnAt:   defaultStopReminderWarnAt,
		},
		Instinct: InstinctValues{
			PersonalPath:     defaultInstinctPersonalPath,
			InheritedPath:    defaultInstinctInheritedPath,
			MinConfidence:    defaultInstinctMinConfidence,
			AutoApprove:      defaultInstinctAutoApprove,
			DecayRate:        defaultInstinctDecayRate,
			MaxInstincts:     defaultInstinctMaxInstincts,
			ClusterThreshold: defaultInstinctClusterThreshold,
		},
	}
}

// getDefaultValue returns the default value for a key as a string.
func getDefaultValue(defaults *Values, key string) string {
	switch key {
	case keyValidateTimeout:
		return strconv.Itoa(defaults.Validate.Timeout)
	case keyValidateCooldown:
		return strconv.Itoa(defaults.Validate.Cooldown)
	case keyNotificationsNtfyTopic:
		return defaults.Notifications.NtfyTopic
	case keyCompactThreshold:
		return strconv.Itoa(defaults.Compact.Threshold)
	case keyCompactReminderInterval:
		return strconv.Itoa(defaults.Compact.ReminderInterval)
	case keyNotifyQuietHoursEnabled:
		return strconv.FormatBool(defaults.Notify.QuietHours.Enabled)
	case keyNotifyQuietHoursStart:
		return defaults.Notify.QuietHours.Start
	case keyNotifyQuietHoursEnd:
		return defaults.Notify.QuietHours.End
	case keyNotifyAudioEnabled:
		return strconv.FormatBool(defaults.Notify.Audio.Enabled)
	case keyNotifyAudioDirectory:
		return defaults.Notify.Audio.Directory
	case keyNotifyDesktopEnabled:
		return strconv.FormatBool(defaults.Notify.Desktop.Enabled)
	case keyObserveEnabled:
		return strconv.FormatBool(defaults.Observe.Enabled)
	case keyObserveMaxFileSizeMB:
		return strconv.Itoa(defaults.Observe.MaxFileSizeMB)
	case keyLearningMinSessionLength:
		return strconv.Itoa(defaults.Learning.MinSessionLength)
	case keyLearningLearnedSkillsPath:
		return defaults.Learning.LearnedSkillsPath
	case keyPreCommitEnabled:
		return strconv.FormatBool(defaults.PreCommit.Enabled)
	case keyPreCommitCommand:
		return defaults.PreCommit.Command
	case keyPackageManagerPreferred:
		return defaults.PackageManager.Preferred
	default:
		v, found, _ := defaults.getExtendedValue(key)
		if found {
			return v
		}
		return ""
	}
}

// allKeys returns all configuration keys in a consistent order.
func allKeys() []string {
	return []string{
		keyValidateTimeout,
		keyValidateCooldown,
		keyNotificationsNtfyTopic,
		keyCompactThreshold,
		keyCompactReminderInterval,
		keyNotifyQuietHoursEnabled,
		keyNotifyQuietHoursStart,
		keyNotifyQuietHoursEnd,
		keyNotifyAudioEnabled,
		keyNotifyAudioDirectory,
		keyNotifyDesktopEnabled,
		keyObserveEnabled,
		keyObserveMaxFileSizeMB,
		keyLearningMinSessionLength,
		keyLearningLearnedSkillsPath,
		keyPreCommitEnabled,
		keyPreCommitCommand,
		keyPackageManagerPreferred,
		keyDriftEnabled,
		keyDriftMinEdits,
		keyDriftThreshold,
		keyStopReminderEnabled,
		keyStopReminderInterval,
		keyStopReminderWarnAt,
		keyInstinctPersonalPath,
		keyInstinctInheritedPath,
		keyInstinctMinConfidence,
		keyInstinctAutoApprove,
		keyInstinctDecayRate,
		keyInstinctMaxInstincts,
		keyInstinctClusterThreshold,
	}
}

#!/usr/bin/env python3
"""
Claude Code Stop Hook - Task Completion Announcer
Plays pre-generated audio clips when tasks are completed
Respects quiet hours (9 PM to 7:30 AM)
"""

import subprocess
import sys
from datetime import datetime
from pathlib import Path


def is_quiet_hours():
    """Check if current time is within quiet hours (9 PM to 7:30 AM)"""
    now = datetime.now()
    current_time = now.time()

    # Define quiet hours
    quiet_start = datetime.strptime("21:00", "%H:%M").time()  # 9 PM
    quiet_end = datetime.strptime("07:30", "%H:%M").time()  # 7:30 AM

    # Check if current time is in quiet hours
    # Note: This handles the case where quiet hours span midnight
    if quiet_start <= quiet_end:
        # Quiet hours don't span midnight (e.g., 6 AM to 8 AM)
        return quiet_start <= current_time <= quiet_end
    else:
        # Quiet hours span midnight (e.g., 9 PM to 7:30 AM)
        return current_time >= quiet_start or current_time <= quiet_end


def play_audio(audio_file):
    """Play audio file using system audio player"""
    try:
        # Use afplay on macOS
        subprocess.run(["afplay", str(audio_file)], check=True)
    except subprocess.CalledProcessError:
        print(f"Failed to play audio: {audio_file}")
    except FileNotFoundError:
        print(
            "Audio player not found. Install afplay or modify script for your system."
        )


def main():
    """Main function for Claude Code stop hook"""
    print("Stop hook triggered!")

    # Check if it's quiet hours
    if is_quiet_hours():
        current_time = datetime.now().strftime("%I:%M %p")
        print(f"🔇 Quiet hours active (9 PM - 7:30 AM). Current time: {current_time}")
        print("Audio notification suppressed.")
        return

    # Get audio directory
    audio_dir = Path(__file__).parent.parent / "audio"

    # Default to task_complete sound
    audio_file = "task_complete.mp3"

    # Override with specific sound if provided as argument
    if len(sys.argv) > 1:
        audio_file = f"{sys.argv[1]}.mp3"

    # Full path to audio file
    audio_path = audio_dir / audio_file

    # Check if audio file exists
    if not audio_path.exists():
        print(f"Audio file not found: {audio_path}")
        print("Run generate_audio_clips.py first to create audio files.")
        sys.exit(1)

    # Play the audio
    play_audio(audio_path)


if __name__ == "__main__":
    main()

export async function getLocalStream(video = true, audio = true): Promise<MediaStream> {
  return navigator.mediaDevices.getUserMedia({ video, audio });
}

export async function getScreenStream(): Promise<MediaStream> {
  return navigator.mediaDevices.getDisplayMedia({ video: true, audio: false });
}

export function stopStream(stream: MediaStream) {
  stream.getTracks().forEach((t) => t.stop());
}

export async function toggleScreenShare(): Promise<MediaStream | null> {
  try {
    return await getScreenStream();
  } catch {
    return null;
  }
}
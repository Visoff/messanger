<script lang="ts">
    import { onMount } from "svelte";

    let streams: MediaStream[] = $state([]);

    type RTCMessage = {
        type: "candidate",
        candidate: RTCIceCandidateInit
    } | {
        type: "answer",
        answer: RTCSessionDescriptionInit
    } | {
        type: "offer",
        offer: RTCSessionDescriptionInit
    }

    let ws: WebSocket | undefined;

    const ip = import.meta.env.VITE_WEBRTC_SERVER_IP

    function signal(data: RTCMessage) {
        if (ws && ws.readyState === 1) {
            ws.send(JSON.stringify(data));
        } else {
            console.error("ws not ready");
        }
    }

    function createPeerConnection() {
        const pc = new RTCPeerConnection({
            iceServers: [
                { urls: `stun:${ip}:3478` },
                {
                    urls: [`turn:${ip}:3478`],
                    username: "username",
                    credential: "password",
                },
            ],
            iceTransportPolicy: "all",
        })

        pc.onicecandidate = (e) => {
            if (e.candidate) {
                signal({ type: "candidate", candidate: e.candidate.toJSON() });
            }
        }

        pc.onnegotiationneeded = async () => {
            const offer = await pc.createOffer()
            await pc.setLocalDescription(offer);
            signal({ type: "offer", offer });
        }

        pc.onsignalingstatechange = (e) => {
            console.log("signaling state", pc.signalingState);
        }

        pc.onicegatheringstatechange = (e) => {
            console.log("ice gathering state", pc.iceGatheringState);
        }

        pc.oniceconnectionstatechange = (e) => {
            console.log("ice connection state", pc.iceConnectionState);
        }

        return pc;
    }

    onMount(() => {
        ws = new WebSocket(
            `${import.meta.env.VITE_WEBSOCKET_SERVER}/room/707006f0-2fda-11f1-8fb6-9e5bddb1d57c`
        );

        ws.onclose = () => {
            console.log("ws closed");
        }

        const pc = createPeerConnection();

        pc.ontrack = (e) => {
            for (const stream of e.streams)  {
                if (stream.getAudioTracks().length > 0) {
                    const audio = document.createElement("audio");
                    audio.srcObject = stream;
                    audio.autoplay = true;
                    document.getElementById("videos")?.appendChild(audio);
                } else {
                    const video = document.createElement("video");
                    video.srcObject = stream;
                    video.autoplay = true;
                    video.controls = true;
                    video.onended = () => {
                        video.srcObject = null;
                        document.getElementById("videos")?.removeChild(video);
                    }
                    document.getElementById("videos")?.appendChild(video);
                }
            }
        }

        const ICECandidateQueue: RTCIceCandidateInit[] = [];

        ws.onmessage = async (e) => {
            const data = JSON.parse(e.data) as RTCMessage;
            if (data.type == "offer") {
                console.log("starting renegotiation");
                await pc.setRemoteDescription(new RTCSessionDescription(data.offer));
                ICECandidateQueue.length = 0;
                const answer = await pc.createAnswer();
                await pc.setLocalDescription(answer);
                signal({ type: "answer", answer });
            } else if (data.type == "answer") {
                await pc.setRemoteDescription(new RTCSessionDescription(data.answer));
                for (const candidate of ICECandidateQueue) {
                    pc.addIceCandidate(new RTCIceCandidate(candidate));
                }
                ICECandidateQueue.length = 0;
            } else if (data.type == "candidate") {
                if (pc.remoteDescription) {
                    pc.addIceCandidate(new RTCIceCandidate(data.candidate));
                } else {
                    ICECandidateQueue.push(data.candidate);
                }
            }
        }

        ws.onopen = () => {
            navigator.mediaDevices.getUserMedia({ video: true, audio: true }).then((u_stream) => {
                const video = document.createElement("video");
                video.srcObject = u_stream;
                video.autoplay = true;
                video.controls = true;
                document.getElementById("videos")?.appendChild(video);

                u_stream.getTracks().forEach((track) => {
                    pc.addTrack(track, u_stream);
                });
            });
            if (ws) {
                ws.onopen = null;
            }
        }
    })
</script>

<div id="videos">
</div>

class GoLiveWebRTC {
    constructor(apiUrl, config) {
        this.apiUrl = apiUrl;
        this.conn = new RTCPeerConnection(config);
        this.conn.addEventListener('connectionstatechange', this.onICEConnectionStateChange);
        this.conn.addEventListener('icecandidate', this.onICECandidate);
    }

    onICECandidate = async e => {
        if (e.candidate !== null) return;
        const res = await fetch(
            this.apiUrl,
            {
                method: 'POST',
                headers: {
                    'content-type': 'application/json'
                }, body: JSON.stringify(e.target.localDescription)
            }
        );
        const sdp = await res.json();
        await e.target.setRemoteDescription(new RTCSessionDescription(sdp));
        console.log('Remote SDP: ', sdp);
    }

    onICEConnectionStateChange = e => {
        console.log('ICE state changed: ', e.target.iceConnectionState);
    }

    connect = async () => {
        const offer = await this.conn.createOffer();
        await this.conn.setLocalDescription(offer);
    }
}

// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract GameRound {
    struct Result {
        string PlayerID;
        uint64 CurrentScore;
        int64 WinScore;
    }

    struct Info {
        string GameID;
        string RoundID;
        string ReplayID;
        string PlayerIDs;
    }

    struct Replay {
        int64 DomainSeparationTag;
        uint64 VRFHeight;
        string HashFunc;
        bytes VRFProof;
        string Address;
        string ReplayCID;
        Info GameInfo;
        Result []GameResults;
    }
}

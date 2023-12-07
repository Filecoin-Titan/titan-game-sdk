// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./Ownable.sol";
import "./GameRound.sol";

contract GameReplayContract is Ownable {
    mapping(string => GameRound.Replay) gameReplayMap;
    string[] replayIDs;


    function saveGameReplay(GameRound.Replay[] memory _replays) public onlyOwner {
        checkParams(_replays);

        for (uint256 i = 0; i < _replays.length; i++) {
            copyReplayToStorage(_replays[i]);
        }
    }

    function checkParams(GameRound.Replay[] memory _replays) internal pure {
        require(_replays.length > 0, "_replays can not empty");

        for (uint256 i = 0; i < _replays.length; i++) {
            checkReplay(_replays[i]);
        }
    }

    function checkReplay(GameRound.Replay memory _replay) internal pure {
        require(_replay.DomainSeparationTag > 0, "Replay.DomainSeparationTag can not 0");
        require(_replay.VRFHeight > 0, "Replay.VRFHeight can not 0");
        require(bytes(_replay.HashFunc).length > 0, "Replay.HashFunc can not empty");
        require(_replay.VRFProof.length > 0, "Replay.VRFProof can not empty");
        require(bytes(_replay.Address).length > 0, "Replay.Address can not empty");
        require(bytes(_replay.ReplayCID).length > 0, "Replay.ReplayCID can not empty");
        require(bytes(_replay.GameInfo.GameID).length > 0, "Replay.GameInfo.GameID can not empty");
        require(bytes(_replay.GameInfo.PlayerIDs).length > 0, "Replay.GameInfo.PlayerIDs can not empty");
        require(bytes(_replay.GameInfo.ReplayID).length > 0, "Replay.GameInfo.ReplayID can not empty");
        require(bytes(_replay.GameInfo.RoundID).length > 0, "Replay.GameInfo.RoundID can not empty");
    }

    function copyReplayToStorage(GameRound.Replay memory _replay) internal {
        GameRound.Replay storage storageReplay = gameReplayMap[_replay.GameInfo.ReplayID];
        if (bytes(storageReplay.GameInfo.ReplayID).length == 0) {
            replayIDs.push(_replay.GameInfo.ReplayID);
        }

        storageReplay.DomainSeparationTag = _replay.DomainSeparationTag;
        storageReplay.VRFHeight = _replay.VRFHeight;
        storageReplay.HashFunc = _replay.HashFunc;
        storageReplay.VRFProof = _replay.VRFProof;
        storageReplay.Address = _replay.Address;
        storageReplay.ReplayCID = _replay.ReplayCID;
        storageReplay.GameInfo = _replay.GameInfo;

        for (uint256 i = 0; i < _replay.GameResults.length; i++) {
            storageReplay.GameResults.push(_replay.GameResults[i]);
        }
    }

    function getGameReplay(string memory _replayID) public view returns (GameRound.Replay memory) {
        GameRound.Replay memory replay = gameReplayMap[_replayID];
        require(replay.VRFProof.length > 0,  string(abi.encodePacked("Game replay not found: ", _replayID)));
        
        return replay;
    }

    function getGameReplayLength() public view returns (uint256) {
       return replayIDs.length;
    }

    function getGameReplayByIndex(uint256  _index) public view returns (GameRound.Replay memory) {
       require(replayIDs.length > _index,  string(abi.encodePacked("out of range: ", _index)));
       
       string memory replayID = replayIDs[_index];
       GameRound.Replay memory replay = gameReplayMap[replayID];
       return replay;
    }
}
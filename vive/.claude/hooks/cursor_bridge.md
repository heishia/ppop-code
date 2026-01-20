---
name: cursor-bridge
trigger: PreToolUse
description: cursor-edit Skill 실행 전 Cursor CLI 연결 상태를 확인하는 hook
---

# Cursor Bridge Hook

cursor-edit Skill 실행 전에 Cursor CLI 연결 상태를 확인한다.

## 트리거 조건

- cursor-edit Skill이 호출될 때
- 코드 수정 작업이 시작될 때

## 확인 항목

1. cursor-agent 명령어 존재 여부
2. Cursor IDE 실행 상태 (선택)
3. 대상 프로젝트 경로 유효성

## 실패 시 동작

- cursor-agent 미설치: 설치 안내 메시지 출력
- 연결 실패: 재시도 또는 사용자에게 확인 요청

## 설정

config/bridges.yaml의 cursor 섹션 참조

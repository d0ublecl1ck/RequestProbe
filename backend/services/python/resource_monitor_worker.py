import hashlib
import json
import mimetypes
import os
import re
import sys
import threading
import time
import traceback
from datetime import datetime, timezone
from pathlib import Path
from urllib.parse import urlparse

try:
    from DrissionPage import ChromiumPage, ChromiumOptions
    IMPORT_ERROR = None
except Exception as exc:  # pragma: no cover
    ChromiumPage = None
    ChromiumOptions = None
    IMPORT_ERROR = str(exc)


def now_iso():
    return datetime.now(timezone.utc).isoformat()


def emit(payload):
    sys.stdout.write(json.dumps(payload, ensure_ascii=False) + "\n")
    sys.stdout.flush()


def response(message_id, ok=True, payload=None, error=""):
    emit({
        "id": message_id,
        "type": "response",
        "ok": ok,
        "payload": payload or {},
        "error": error,
    })


def event(event_type, payload=None, message=""):
    emit({
        "type": event_type,
        "payload": payload or {},
        "message": message,
    })


def sanitize_name(name):
    cleaned = re.sub(r"[^\w.\-]+", "-", name, flags=re.UNICODE).strip("-.")
    return cleaned or "resource"


def ensure_bytes(body):
    if body is None:
        return b""
    if isinstance(body, bytes):
        return body
    if isinstance(body, bytearray):
        return bytes(body)
    if isinstance(body, str):
        return body.encode("utf-8", errors="ignore")
    return json.dumps(body, ensure_ascii=False, sort_keys=True).encode("utf-8", errors="ignore")


def normalize_extension(ext):
    ext = (ext or "").strip().lower().lstrip(".")
    return ext


def infer_extension(url, mime_type):
    path = urlparse(url).path
    suffix = Path(path).suffix.lower().lstrip(".")
    if suffix:
        return suffix

    mime_guess = mimetypes.guess_extension((mime_type or "").split(";")[0].strip()) or ""
    return mime_guess.lower().lstrip(".")


def infer_filename(url, extension, digest):
    path_name = Path(urlparse(url).path).name
    if not path_name:
        path_name = f"resource.{extension}" if extension else "resource"
    safe_name = sanitize_name(path_name)
    stem = Path(safe_name).stem or "resource"
    suffix = Path(safe_name).suffix
    if not suffix and extension:
        suffix = f".{extension}"
    return f"{stem}-{digest[:10]}{suffix}"


class ResourceMonitorWorker:
    def __init__(self):
        self.lock = threading.RLock()
        self.page = None
        self.listener_thread = None
        self.navigation_thread = None
        self.listener_stop = threading.Event()
        self.current_task = None
        self.extensions = set()
        self.resources = {}

    def start_listener_thread(self):
        if self.listener_thread and self.listener_thread.is_alive():
            return
        self.listener_stop.clear()
        self.listener_thread = threading.Thread(target=self.listen_loop, daemon=True)
        self.listener_thread.start()

    def close_browser(self):
        page = self.page
        self.page = None
        self.navigation_thread = None
        if page is None:
            return
        try:
            page.listen.stop()
        except Exception:
            pass
        try:
            page.quit()
        except Exception:
            try:
                page.close()
            except Exception:
                pass

    def public_task(self):
        if not self.current_task:
            return None
        payload = dict(self.current_task)
        payload["resources"] = [self.public_resource(item) for item in self.resources.values()]
        return payload

    def public_resource(self, item):
        return {
            "id": item["id"],
            "url": item["url"],
            "extension": item["extension"],
            "hash": item["hash"],
            "mimeType": item.get("mimeType", ""),
            "statusCode": item.get("statusCode", 0),
            "suggestedFileName": item["suggestedFileName"],
            "size": item["size"],
            "downloaded": item.get("downloaded", False),
            "downloadedPath": item.get("downloadedPath", ""),
            "firstSeenAt": item["firstSeenAt"],
            "lastSeenAt": item["lastSeenAt"],
        }

    def update_status(self, status, last_error=""):
        if not self.current_task:
            return
        self.current_task["status"] = status
        self.current_task["updatedAt"] = now_iso()
        self.current_task["lastError"] = last_error
        event("task_updated", self.public_task())

    def start_task(self, payload):
        if IMPORT_ERROR:
            raise RuntimeError(f"DrissionPage 不可用: {IMPORT_ERROR}")

        url = (payload.get("url") or "").strip()
        task_id = payload.get("taskId") or ""
        extensions = {normalize_extension(item) for item in payload.get("extensions") or []}
        extensions.discard("")
        download_dir = payload.get("downloadDir") or ""

        if not task_id:
            raise ValueError("taskId 不能为空")
        if not extensions:
            raise ValueError("至少选择一个文件后缀")
        if not download_dir:
            raise ValueError("downloadDir 不能为空")

        with self.lock:
            self.listener_stop.set()
            self.close_browser()
            self.resources = {}
            self.extensions = extensions
            os.makedirs(download_dir, exist_ok=True)
            browser_workspace = os.path.join(download_dir, ".browser")
            os.makedirs(browser_workspace, exist_ok=True)
            self.current_task = {
                "taskId": task_id,
                "url": url,
                "status": "running",
                "selectedExtensions": sorted(extensions),
                "downloadDir": download_dir,
                "createdAt": now_iso(),
                "updatedAt": now_iso(),
                "lastError": "",
            }
            options = ChromiumOptions(read_file=False)
            options.auto_port()
            options.new_env(True)
            options.set_download_path(download_dir)
            options.set_tmp_path(browser_workspace)
            self.page = ChromiumPage(options)
            self.page.listen.start()
            self.start_listener_thread()
            if url:
                self.navigation_thread = threading.Thread(
                    target=self.navigate_to_url,
                    args=(url,),
                    daemon=True,
                )
                self.navigation_thread.start()
            event("task_updated", self.public_task())
            return self.public_task()

    def navigate_to_url(self, url):
        try:
            page = self.page
            if page is None:
                return
            page.get(url)
        except Exception:
            with self.lock:
                if self.current_task and self.current_task.get("status") != "ended":
                    self.current_task["lastError"] = f"页面导航失败: {traceback.format_exc().splitlines()[-1]}"
                    self.current_task["updatedAt"] = now_iso()
                    event("task_updated", self.public_task())

    def pause_task(self):
        with self.lock:
            if not self.page or not self.current_task:
                raise RuntimeError("当前没有运行中的任务")
            self.page.listen.pause()
            self.update_status("paused")
            return self.public_task()

    def resume_task(self):
        with self.lock:
            if not self.page or not self.current_task:
                raise RuntimeError("当前没有可恢复的任务")
            self.page.listen.resume()
            self.update_status("running")
            return self.public_task()

    def end_task(self):
        with self.lock:
            if not self.current_task:
                raise RuntimeError("当前没有任务可以结束")
            self.listener_stop.set()
            self.close_browser()
            self.update_status("ended")
            return self.public_task()

    def current_state(self):
        with self.lock:
            return self.public_task()

    def handle_packet(self, packet):
        response_obj = getattr(packet, "response", None)
        if response_obj is None:
            return

        url = getattr(packet, "url", "") or ""
        if not url:
            return

        try:
            headers = getattr(response_obj, "headers", None) or {}
        except Exception:
            headers = {}
        mime_type = headers.get("content-type", "")
        extension = infer_extension(url, mime_type)
        if extension not in self.extensions:
            return

        try:
            body = ensure_bytes(getattr(response_obj, "body", None))
        except Exception:
            return
        if not body:
            return

        digest = hashlib.sha256(body).hexdigest()
        with self.lock:
            existing = self.resources.get(digest)
            if existing:
                existing["lastSeenAt"] = now_iso()
                return

            try:
                status_code = getattr(response_obj, "status", 0) or 0
            except Exception:
                status_code = 0
            item = {
                "id": digest,
                "hash": digest,
                "url": url,
                "extension": extension,
                "mimeType": mime_type,
                "statusCode": status_code,
                "suggestedFileName": infer_filename(url, extension, digest),
                "size": len(body),
                "downloaded": False,
                "downloadedPath": "",
                "firstSeenAt": now_iso(),
                "lastSeenAt": now_iso(),
                "_bytes": body,
            }
            self.resources[digest] = item
            if self.current_task:
                self.current_task["updatedAt"] = now_iso()

        event("resource_detected", {"task": self.public_task(), "resource": self.public_resource(item)})

    def download_resources(self, payload):
        resource_ids = payload.get("resourceIds") or []
        if not resource_ids:
            raise ValueError("未选择任何资源")

        with self.lock:
            if not self.current_task:
                raise RuntimeError("当前没有活动任务")
            download_dir = self.current_task["downloadDir"]

        downloaded_ids = []
        skipped_ids = []
        downloaded_entries = []

        for resource_id in resource_ids:
            with self.lock:
                item = self.resources.get(resource_id)
                if not item:
                    skipped_ids.append(resource_id)
                    continue
                file_name = item["suggestedFileName"]
                target_path = os.path.join(download_dir, file_name)
                if item.get("downloaded") and item.get("downloadedPath") and os.path.exists(item["downloadedPath"]):
                    skipped_ids.append(resource_id)
                    continue
                if os.path.exists(target_path):
                    item["downloaded"] = True
                    item["downloadedPath"] = target_path
                    skipped_ids.append(resource_id)
                    continue
                data = item.get("_bytes") or b""

            with open(target_path, "wb") as fh:
                fh.write(data)

            with self.lock:
                item["downloaded"] = True
                item["downloadedPath"] = target_path
                item["lastSeenAt"] = now_iso()
                downloaded_ids.append(resource_id)
                downloaded_entries.append(self.public_resource(item))

        result = {
            "taskId": self.current_task["taskId"],
            "downloadDir": self.current_task["downloadDir"],
            "downloadedIds": downloaded_ids,
            "skippedIds": skipped_ids,
            "downloadedEntries": downloaded_entries,
        }
        event("resources_downloaded", result)
        return result

    def listen_loop(self):
        while not self.listener_stop.is_set():
            with self.lock:
                page = self.page
            if page is None:
                time.sleep(0.1)
                continue

            try:
                packets = page.listen.steps(timeout=1, gap=1)
                if packets is False:
                    continue
                for item in packets:
                    if item is False:
                        break
                    batch = item if isinstance(item, list) else [item]
                    for packet in batch:
                        self.handle_packet(packet)
            except Exception:
                if self.listener_stop.is_set():
                    break
                event("worker_log", {"trace": traceback.format_exc()})
                time.sleep(0.2)


def main():
    worker = ResourceMonitorWorker()
    for raw_line in sys.stdin:
        line = raw_line.strip()
        if not line:
            continue
        try:
            message = json.loads(line)
            message_id = message.get("id")
            command_type = message.get("type")
            payload = message.get("payload") or {}

            if command_type == "start_task":
                result = worker.start_task(payload)
            elif command_type == "pause_task":
                result = worker.pause_task()
            elif command_type == "resume_task":
                result = worker.resume_task()
            elif command_type == "end_task":
                result = worker.end_task()
            elif command_type == "get_state":
                result = worker.current_state() or {}
            elif command_type == "download_resources":
                result = worker.download_resources(payload)
            elif command_type == "ping":
                result = {"pong": True}
            else:
                raise ValueError(f"未知命令: {command_type}")

            response(message_id, ok=True, payload=result)
        except Exception as exc:
            response(message.get("id") if "message" in locals() and isinstance(message, dict) else "", ok=False, error=str(exc))


if __name__ == "__main__":
    main()

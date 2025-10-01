import asyncio
import os

from typing import List, Dict, Awaitable
from ling_chat.core.TTS.tts_provider import TTS
from ling_chat.core.logger import logger

class VoiceMaker:
    def __init__(self) -> None:
        self.tts_provider = TTS()
        self.model_name = ""
        self.speaker_id = 4
        self.tts_type = ""
        self.lang = "ja"  # 默认语言为日语

        # 初始化语音合成器可用状态
        self.sva_available = False
        self.sbv2_available = False
        self.bv2_available = False
        self.sbv2api_available = False
        self.gsv_available = False
        self.aivis_available = False

    def check_tts_availability(self, tts_settings: dict[str, str]) -> None:
        """检查TTS配置可用性，设置各语音合成器状态"""
        
        def _is_valid(value: str) -> bool:
            """检查字符串是否有效（非空且非空格）"""
            return value is not None and value.strip() != ""
        
        # 检查SVA配置
        sva_speaker_id = tts_settings.get("sva_speaker_id", "")
        self.sva_available = _is_valid(sva_speaker_id)
        
        # 检查SBV2配置
        sbv2_speaker_id = tts_settings.get("sbv2_speaker_id", "")
        sbv2_name = tts_settings.get("sbv2_name", "")
        self.sbv2_available = _is_valid(sbv2_speaker_id) and _is_valid(sbv2_name)
        
        # 检查BV2配置
        bv2_speaker_id = tts_settings.get("bv2_speaker_id", "")
        self.bv2_available = _is_valid(bv2_speaker_id)
        
        # 检查SBV2API配置
        sbv2api_name = tts_settings.get("sbv2api_name", "")
        sbv2api_speaker_id = tts_settings.get("sbv2api_speaker_id", "")
        self.sbv2api_available = _is_valid(sbv2api_name) and _is_valid(sbv2api_speaker_id)
        
        # 检查GSV配置
        gsv_voice_filename = tts_settings.get("gsv_voice_filename", "")
        gsv_voice_text = tts_settings.get("gsv_voice_text", "")
        self.gsv_available = _is_valid(gsv_voice_filename) and _is_valid(gsv_voice_text)
        
        # 检查AIVIS配置
        aivis_model_uuid = tts_settings.get("aivis_model_uuid", "")
        self.aivis_available = _is_valid(aivis_model_uuid)

    def set_tts_settings(self, tts_settings: dict[str,str], name: str) -> None:
        """获取可用的TTS配置并且进行基础配置"""
        try:
            # 先检查所有TTS配置的可用性
            logger.debug("开始验证TTS配置可用性")
            self.check_tts_availability(tts_settings)
            
            # 根据当前设置的TTS类型进行初始化
            if self.tts_type == "sva-vits" and self.sva_available:
                self.tts_provider.init_sva_adapter(speaker_id=int(tts_settings["sva_speaker_id"]))
            elif self.tts_type == "sbv2" and self.sbv2_available:
              	self.tts_provider.init_sbv2_adapter(speaker_id=int(tts_settings["sbv2_speaker_id"]), 
                                                    model_name=tts_settings["sbv2_name"])
            elif self.tts_type == "sva-bv2" and self.bv2_available:
                self.tts_provider.init_bv2_adapter(speaker_id=int(tts_settings["bv2_speaker_id"]))
            elif self.tts_type == "sbv2api" and self.sbv2api_available:
                self.tts_provider.init_sbv2api_adapter(model_name=tts_settings["sbv2api_name"],
                                                       speaker_id=int(tts_settings["sbv2api_speaker_id"]))
            elif self.tts_type == "gsv" and self.gsv_available:
                # 优先使用环境变量定义的语音文件
                if os.environ.get("GPT_SOVITS_REF_AUDIO", "") == "":
                    self.tts_provider.init_gsv_adapter(ref_audio_path=tts_settings["gsv_voice_filename"],
                                                       prompt_text=tts_settings["gsv_voice_text"])
                else:
                    self.tts_provider.init_gsv_adapter(ref_audio_path=os.environ.get("GPT_SOVITS_REF_AUDIO", ""),
                                                       prompt_text=os.environ.get("GPT_SOVITS_PROMPT_TEXT", ""))
                    logger.warning("你正在使用环境变量中的GPT-SoVITS配置")
            elif self.tts_type == "aivis" and self.aivis_available:
                self.tts_provider.init_aivis_adapter(model_uuid=tts_settings["aivis_model_uuid"])
            else:
                logger.warning(f"你的环境变量中TTS设置有误，此角色{name}不支持{self.tts_type}，将使用角色卡的默认语音合成器！")
                raise ValueError
        except KeyError as e:
            logger.error(f"当前角色卡{name}的TTS设置出错，问题是：{e}")

    def set_tts(self, tts_type: str, tts_settings: dict[str,str], name: str) -> None:
        """设置默认的TTS类型"""
        available_tts_types = ("sva-bv2", "gsv", "sbv2", "sva-vits", "sbv2api","aivis")
        try:
            if os.environ.get("TTS_TYPE", "") in available_tts_types:
                self.tts_type = os.environ.get("TTS_TYPE", "")
                self.set_tts_settings(tts_settings, name)
            else:
                logger.warning("你的环境变量中未设置TTS类型（或是设置错误），将使用角色卡的默认语音合成器！")
                if tts_type in available_tts_types:
                    self.tts_type = tts_type
                    self.set_tts_settings(tts_settings, name)
        except ValueError:
            if tts_type in available_tts_types:
                self.tts_type = tts_type
                self.set_tts_settings(tts_settings, name)
            else:
                logger.error(f"角色卡中有未知的TTS类型: {tts_type}，请联系角色卡制造者。")
    
    def set_lang(self, lang: str) -> None:
        """设置语言"""
        if lang not in ["ja", "zh"]:
            raise ValueError(f"不支持的语言: {lang}")
        self.lang = lang
    
    async def generate_voice_files(self, segments: List[Dict[str, str]]):
        """生成语音文件"""
        tasks: List[Awaitable[str | None]] = []
        logger.debug(f"生成语音文件: {segments}")
        for seg in segments:
            if self.lang == "ja":
                if seg["japanese_text"]:
                    task = self.tts_provider.generate_voice(seg["japanese_text"], 
                                                            seg["voice_file"], 
                                                            tts_type=self.tts_type, 
                                                            lang="ja")
                    tasks.append(task)
                elif seg["following_text"] and not seg.get("japanese_text"):
                    logger.warning(f"片段 {seg['index']} 没有日语文本，跳过语音生成")
            elif self.lang == "zh":
                if seg["following_text"]:
                    task = self.tts_provider.generate_voice(seg["following_text"], 
                                                            seg["voice_file"], 
                                                            tts_type=self.tts_type, 
                                                            lang="zh")
                    tasks.append(task)
                else:
                    logger.warning(f"片段 {seg['index']} 没有中文文本，跳过语音生成\n"
                                   f"Tips：要真出现这情况，你应该检查LLM是否正常输出。")
        if tasks:
            await asyncio.gather(*tasks)

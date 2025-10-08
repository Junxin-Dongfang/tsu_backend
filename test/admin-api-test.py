#!/usr/bin/env python3
"""
Admin API 接口自动化测试脚本 (Python 版本)
功能: 全面测试 Admin 服务所有接口，生成详细报告
使用: python3 admin-api-test.py [--url URL] [--username USER] [--password PASS]
"""

import requests
import json
import argparse
import sys
import time
from datetime import datetime
from typing import Dict, List, Optional, Tuple
from dataclasses import dataclass, field
from enum import Enum


class TestStatus(Enum):
    """测试状态枚举"""
    PASSED = "PASSED"
    FAILED = "FAILED"
    SKIPPED = "SKIPPED"
    BLOCKED = "BLOCKED"


@dataclass
class TestResult:
    """测试结果数据类"""
    name: str
    status: TestStatus
    http_code: Optional[int] = None
    response_time: float = 0.0
    error_message: str = ""
    request_url: str = ""
    response_data: Optional[Dict] = None


@dataclass
class TestSuite:
    """测试套件"""
    name: str
    results: List[TestResult] = field(default_factory=list)
    
    @property
    def total(self) -> int:
        return len(self.results)
    
    @property
    def passed(self) -> int:
        return sum(1 for r in self.results if r.status == TestStatus.PASSED)
    
    @property
    def failed(self) -> int:
        return sum(1 for r in self.results if r.status == TestStatus.FAILED)
    
    @property
    def skipped(self) -> int:
        return sum(1 for r in self.results if r.status == TestStatus.SKIPPED)
    
    @property
    def pass_rate(self) -> float:
        return (self.passed / self.total * 100) if self.total > 0 else 0.0


class Colors:
    """终端颜色"""
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    MAGENTA = '\033[0;35m'
    CYAN = '\033[0;36m'
    BOLD = '\033[1m'
    NC = '\033[0m'


class AdminAPITester:
    """Admin API 测试器"""
    
    def __init__(self, base_url: str, username: str, password: str):
        self.base_url = base_url.rstrip('/')
        self.api_base = f"{self.base_url}/api/v1"
        self.username = username
        self.password = password
        self.auth_token: Optional[str] = None
        self.session = requests.Session()
        self.test_suites: List[TestSuite] = []
        self.start_time = datetime.now()
        
    def print_colored(self, message: str, color: str = Colors.NC):
        """打印彩色信息"""
        print(f"{color}{message}{Colors.NC}")
        
    def print_header(self, title: str):
        """打印测试标题"""
        print(f"\n{Colors.BLUE}{'=' * 60}{Colors.NC}")
        print(f"{Colors.BOLD}{title}{Colors.NC}")
        print(f"{Colors.BLUE}{'=' * 60}{Colors.NC}\n")
        
    def http_request(
        self,
        method: str,
        endpoint: str,
        description: str,
        data: Optional[Dict] = None,
        auth_required: bool = True,
        expected_codes: List[int] = None
    ) -> TestResult:
        """
        执行 HTTP 请求并返回测试结果
        
        Args:
            method: HTTP 方法
            endpoint: API 端点
            description: 测试描述
            data: 请求数据
            auth_required: 是否需要认证
            expected_codes: 预期的 HTTP 状态码列表
        """
        if expected_codes is None:
            expected_codes = [200, 201]
            
        url = f"{self.base_url}{endpoint}"
        headers = {"Content-Type": "application/json"}
        
        if auth_required:
            if not self.auth_token:
                return TestResult(
                    name=description,
                    status=TestStatus.BLOCKED,
                    error_message="未设置认证 Token",
                    request_url=url
                )
            headers["Authorization"] = f"Bearer {self.auth_token}"
        
        try:
            start = time.time()
            response = self.session.request(
                method=method,
                url=url,
                headers=headers,
                json=data,
                timeout=10
            )
            response_time = time.time() - start
            
            try:
                response_data = response.json()
            except:
                response_data = {"raw": response.text}
            
            status = TestStatus.PASSED if response.status_code in expected_codes else TestStatus.FAILED
            
            result = TestResult(
                name=description,
                status=status,
                http_code=response.status_code,
                response_time=response_time,
                request_url=url,
                response_data=response_data
            )
            
            if status == TestStatus.FAILED:
                result.error_message = f"预期状态码 {expected_codes}, 实际 {response.status_code}"
            
            # 打印结果
            if status == TestStatus.PASSED:
                self.print_colored(
                    f"[✓ PASS] {description} - HTTP {response.status_code} ({response_time:.2f}s)",
                    Colors.GREEN
                )
            else:
                self.print_colored(
                    f"[✗ FAIL] {description} - HTTP {response.status_code} - {result.error_message}",
                    Colors.RED
                )
            
            return result
            
        except requests.exceptions.RequestException as e:
            self.print_colored(
                f"[✗ FAIL] {description} - 请求异常: {str(e)}",
                Colors.RED
            )
            return TestResult(
                name=description,
                status=TestStatus.FAILED,
                error_message=str(e),
                request_url=url
            )
    
    def login(self) -> bool:
        """登录并获取 Token"""
        self.print_header("认证测试")
        
        login_data = {
            "identifier": self.username,  # 使用 identifier 而不是 username
            "password": self.password
        }
        
        result = self.http_request(
            method="POST",
            endpoint="/api/v1/auth/login",
            description="用户登录",
            data=login_data,
            auth_required=False
        )
        
        if result.status == TestStatus.PASSED and result.response_data:
            # 尝试从不同的响应格式中提取 token
            token = (
                result.response_data.get('data', {}).get('session_token') or
                result.response_data.get('data', {}).get('token') or
                result.response_data.get('token') or
                result.response_data.get('session_token') or
                result.response_data.get('access_token')
            )
            
            if token:
                self.auth_token = token
                self.print_colored(f"Token: {token[:30]}...", Colors.CYAN)
                return True
            else:
                self.print_colored("⚠ 登录成功但未找到 Token", Colors.YELLOW)
                print(json.dumps(result.response_data, indent=2))
                return False
        else:
            self.print_colored("⚠ 登录失败", Colors.RED)
            return False
    
    # ==================== 测试套件 ====================
    
    def test_system_health(self):
        """测试系统健康"""
        self.print_header("1. 系统健康检查")
        suite = TestSuite(name="系统健康检查")
        
        suite.results.append(self.http_request(
            "GET", "/health", "健康检查接口", auth_required=False
        ))
        
        suite.results.append(self.http_request(
            "GET", "/swagger/index.html", "Swagger 文档可访问性",
            auth_required=False, expected_codes=[200, 301, 302]
        ))
        
        self.test_suites.append(suite)
    
    def test_authentication(self):
        """测试认证流程"""
        suite = TestSuite(name="认证流程")
        
        suite.results.append(self.http_request(
            "GET", "/api/v1/admin/users/me", "获取当前用户信息"
        ))
        
        self.test_suites.append(suite)
    
    def test_user_management(self):
        """测试用户管理"""
        self.print_header("2. 用户管理测试")
        suite = TestSuite(name="用户管理")
        
        # 获取用户列表
        result = self.http_request(
            "GET", "/api/v1/admin/users?page=1&page_size=10", "获取用户列表"
        )
        suite.results.append(result)
        
        # 尝试从用户列表中获取第一个用户的 ID
        user_id = None
        if result.status == TestStatus.PASSED and result.response_data:
            items = result.response_data.get('data', {}).get('items', [])
            if items:
                user_id = items[0].get('id')
        
        # 如果有用户 ID，测试获取用户详情
        if user_id:
            suite.results.append(self.http_request(
                "GET", f"/api/v1/admin/users/{user_id}", 
                f"获取用户详情 (ID: {user_id})"
            ))
        else:
            # 没有用户，跳过测试
            self.print_colored("⊘ 跳过用户详情测试 - 暂无用户数据", Colors.YELLOW)
        
        self.test_suites.append(suite)
    
    def test_rbac(self):
        """测试 RBAC 权限系统"""
        self.print_header("3. RBAC 权限系统测试")
        suite = TestSuite(name="RBAC权限系统")
        
        # 角色
        suite.results.append(self.http_request(
            "GET", "/api/v1/admin/roles?page=1&page_size=10", "获取角色列表"
        ))
        
        # 权限
        suite.results.append(self.http_request(
            "GET", "/api/v1/admin/permissions?page=1&page_size=10", "获取权限列表"
        ))
        
        suite.results.append(self.http_request(
            "GET", "/api/v1/admin/permission-groups", "获取权限组列表"
        ))
        
        # 获取实际存在的用户 ID
        users_result = self.http_request(
            "GET", "/api/v1/admin/users?page=1&page_size=1", "获取用户（用于权限测试）"
        )
        suite.results.append(users_result)
        
        user_id = None
        if users_result.status == TestStatus.PASSED and users_result.response_data:
            items = users_result.response_data.get('data', {}).get('items', [])
            if items:
                user_id = items[0].get('id')
        
        # 如果有用户，测试用户权限相关接口
        if user_id:
            suite.results.append(self.http_request(
                "GET", f"/api/v1/admin/users/{user_id}/roles", 
                f"获取用户角色 (ID: {user_id})"
            ))
            
            suite.results.append(self.http_request(
                "GET", f"/api/v1/admin/users/{user_id}/permissions", 
                f"获取用户权限 (ID: {user_id})"
            ))
        else:
            self.print_colored("⊘ 跳过用户权限测试 - 暂无用户数据", Colors.YELLOW)
        
        self.test_suites.append(suite)
    
    def test_basic_game_config(self):
        """测试基础游戏配置"""
        self.print_header("4. 基础游戏配置测试")
        suite = TestSuite(name="基础游戏配置")
        
        config_types = [
            ("classes", "职业"),
            ("skill-categories", "技能分类"),
            ("action-categories", "动作分类"),
            ("damage-types", "伤害类型"),
            ("hero-attribute-types", "英雄属性类型"),
            ("tags", "标签"),
            ("action-flags", "动作标记"),
        ]
        
        for endpoint, name in config_types:
            suite.results.append(self.http_request(
                "GET", f"/api/v1/admin/{endpoint}?page=1&page_size=10",
                f"获取{name}列表"
            ))
        
        # 标签关系使用不同的路由结构，测试基于实体的查询
        # 先获取一个标签 ID 用于测试
        tags_result = self.http_request(
            "GET", "/api/v1/admin/tags?page=1&page_size=1",
            "获取标签（用于关系测试）"
        )
        suite.results.append(tags_result)
        
        if tags_result.status == TestStatus.PASSED and tags_result.response_data:
            items = tags_result.response_data.get('data', {}).get('items', [])
            if items:
                tag_id = items[0].get('id')
                suite.results.append(self.http_request(
                    "GET", f"/api/v1/admin/tags/{tag_id}/entities",
                    f"获取标签关联的实体 (Tag ID: {tag_id})"
                ))
            else:
                self.print_colored("⊘ 跳过标签关系测试 - 暂无标签数据", Colors.YELLOW)
        
        self.test_suites.append(suite)
    
    def test_metadata_definitions(self):
        """测试元数据定义"""
        self.print_header("5. 元数据定义测试")
        suite = TestSuite(name="元数据定义")
        
        definitions = [
            ("effect-type-definitions", "效果类型定义"),
            ("formula-variables", "公式变量"),
            ("range-config-rules", "范围配置规则"),
            ("action-type-definitions", "动作类型定义"),
        ]
        
        for endpoint, name in definitions:
            suite.results.append(self.http_request(
                "GET", f"/api/v1/admin/metadata/{endpoint}?page=1&page_size=10",
                f"获取{name}列表 (分页)"
            ))
            
            suite.results.append(self.http_request(
                "GET", f"/api/v1/admin/metadata/{endpoint}/all",
                f"获取所有{name}"
            ))
        
        self.test_suites.append(suite)
    
    def test_skill_system(self):
        """测试技能系统"""
        self.print_header("6. 技能系统测试")
        suite = TestSuite(name="技能系统")
        
        # 获取技能列表
        result = self.http_request(
            "GET", "/api/v1/admin/skills?page=1&page_size=10", "获取技能列表"
        )
        suite.results.append(result)
        
        # 尝试获取第一个技能的详情
        if result.status == TestStatus.PASSED and result.response_data:
            items = result.response_data.get('data', {}).get('items', [])
            if items:
                skill_id = items[0].get('id')
                suite.results.append(self.http_request(
                    "GET", f"/api/v1/admin/skills/{skill_id}",
                    f"获取技能详情 (ID: {skill_id})"
                ))
                
                suite.results.append(self.http_request(
                    "GET", f"/api/v1/admin/skills/{skill_id}/level-configs",
                    f"获取技能等级配置 (ID: {skill_id})"
                ))
                
                suite.results.append(self.http_request(
                    "GET", f"/api/v1/admin/skills/{skill_id}/unlock-actions",
                    f"获取技能解锁动作 (ID: {skill_id})"
                ))
        
        self.test_suites.append(suite)
    
    def test_effect_system(self):
        """测试效果系统"""
        self.print_header("7. 效果系统测试")
        suite = TestSuite(name="效果系统")
        
        # Effects
        result = self.http_request(
            "GET", "/api/v1/admin/effects?page=1&page_size=10", "获取效果列表"
        )
        suite.results.append(result)
        
        # Buffs
        buff_result = self.http_request(
            "GET", "/api/v1/admin/buffs?page=1&page_size=10", "获取Buff列表"
        )
        suite.results.append(buff_result)
        
        # 测试 Buff 详情
        if buff_result.status == TestStatus.PASSED and buff_result.response_data:
            items = buff_result.response_data.get('data', {}).get('items', [])
            if items:
                buff_id = items[0].get('id')
                suite.results.append(self.http_request(
                    "GET", f"/api/v1/admin/buffs/{buff_id}",
                    f"获取Buff详情 (ID: {buff_id})"
                ))
                
                suite.results.append(self.http_request(
                    "GET", f"/api/v1/admin/buffs/{buff_id}/effects",
                    f"获取Buff关联的效果 (ID: {buff_id})"
                ))
        
        self.test_suites.append(suite)
    
    def test_action_system(self):
        """测试动作系统"""
        self.print_header("8. 动作系统测试")
        suite = TestSuite(name="动作系统")
        
        result = self.http_request(
            "GET", "/api/v1/admin/actions?page=1&page_size=10", "获取动作列表"
        )
        suite.results.append(result)
        
        # 测试动作详情
        if result.status == TestStatus.PASSED and result.response_data:
            items = result.response_data.get('data', {}).get('items', [])
            if items:
                action_id = items[0].get('id')
                suite.results.append(self.http_request(
                    "GET", f"/api/v1/admin/actions/{action_id}",
                    f"获取动作详情 (ID: {action_id})"
                ))
                
                suite.results.append(self.http_request(
                    "GET", f"/api/v1/admin/actions/{action_id}/effects",
                    f"获取动作关联的效果 (ID: {action_id})"
                ))
        
        self.test_suites.append(suite)
    
    def test_error_handling(self):
        """测试错误处理"""
        self.print_header("9. 错误处理测试")
        suite = TestSuite(name="错误处理")
        
        # 测试 404
        suite.results.append(self.http_request(
            "GET", "/api/v1/admin/skills/999999",
            "404错误 - 访问不存在的资源",
            expected_codes=[404]
        ))
        
        # 测试无效分页
        suite.results.append(self.http_request(
            "GET", "/api/v1/admin/skills?page=-1&page_size=0",
            "参数验证 - 无效的分页参数",
            expected_codes=[200, 400]  # 可能返回 200 并使用默认值
        ))
        
        # 测试未认证访问 - 使用新的请求不带任何认证信息
        old_token = self.auth_token
        self.auth_token = None
        try:
            start = time.time()
            # 使用requests.get而不是session，确保没有任何缓存的认证信息
            response = requests.get(
                url=f"{self.base_url}/api/v1/admin/users",
                timeout=10
            )
            response_time = time.time() - start
            
            status = TestStatus.PASSED if response.status_code in [401, 403] else TestStatus.FAILED
            result = TestResult(
                name="401/403错误 - 未认证访问受保护接口",
                status=status,
                http_code=response.status_code,
                response_time=response_time,
                request_url=f"{self.base_url}/api/v1/admin/users"
            )
            if status == TestStatus.FAILED:
                result.error_message = f"预期状态码 [401, 403], 实际 {response.status_code}"
            
            suite.results.append(result)
            
            if status == TestStatus.PASSED:
                self.print_colored(
                    f"[✓ PASS] 401/403错误 - 未认证访问受保护接口 - HTTP {response.status_code} ({response_time:.2f}s)",
                    Colors.GREEN
                )
            else:
                self.print_colored(
                    f"[✗ FAIL] 401/403错误 - 未认证访问受保护接口 - {result.error_message}",
                    Colors.RED
                )
        except Exception as e:
            suite.results.append(TestResult(
                name="401/403错误 - 未认证访问受保护接口",
                status=TestStatus.FAILED,
                error_message=str(e),
                request_url=f"{self.base_url}/api/v1/admin/users"
            ))
        finally:
            self.auth_token = old_token
        
        # 测试无效 Token
        self.auth_token = "invalid_token_12345"
        suite.results.append(self.http_request(
            "GET", "/api/v1/admin/users",
            "401/403错误 - 无效Token访问",
            expected_codes=[401, 403]
        ))
        self.auth_token = old_token
        
        self.test_suites.append(suite)
    
    # ==================== 报告生成 ====================
    
    def generate_console_report(self):
        """生成控制台报告"""
        self.print_header("测试报告")
        
        total_tests = sum(suite.total for suite in self.test_suites)
        total_passed = sum(suite.passed for suite in self.test_suites)
        total_failed = sum(suite.failed for suite in self.test_suites)
        total_skipped = sum(suite.skipped for suite in self.test_suites)
        pass_rate = (total_passed / total_tests * 100) if total_tests > 0 else 0
        
        duration = datetime.now() - self.start_time
        
        print(f"\n测试开始时间: {self.start_time.strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"测试结束时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"测试耗时: {duration.total_seconds():.2f} 秒\n")
        
        print("━" * 60)
        print("  总体统计")
        print("━" * 60)
        print(f"  总测试数:   {total_tests}")
        self.print_colored(f"  通过:       {total_passed}", Colors.GREEN)
        self.print_colored(f"  失败:       {total_failed}", Colors.RED)
        self.print_colored(f"  跳过:       {total_skipped}", Colors.YELLOW)
        print(f"  通过率:     {pass_rate:.1f}%")
        print("━" * 60)
        
        # 各套件详情
        print("\n━" * 60)
        print("  测试套件详情")
        print("━" * 60)
        for suite in self.test_suites:
            status_icon = "✓" if suite.failed == 0 else "✗"
            color = Colors.GREEN if suite.failed == 0 else Colors.RED
            self.print_colored(
                f"  {status_icon} {suite.name}: {suite.passed}/{suite.total} 通过 ({suite.pass_rate:.1f}%)",
                color
            )
        print("━" * 60)
        
        # 失败用例
        if total_failed > 0:
            print("\n━" * 60)
            print("  失败用例详情")
            print("━" * 60)
            for suite in self.test_suites:
                failed_results = [r for r in suite.results if r.status == TestStatus.FAILED]
                if failed_results:
                    self.print_colored(f"\n  [{suite.name}]", Colors.YELLOW)
                    for result in failed_results:
                        print(f"    ✗ {result.name}")
                        print(f"      URL: {result.request_url}")
                        print(f"      状态码: {result.http_code}")
                        print(f"      错误: {result.error_message}")
            print("━" * 60)
        
        # 最终结论
        print()
        if total_failed == 0:
            self.print_colored("✓ 所有测试通过！", Colors.GREEN)
        else:
            self.print_colored(f"⚠ 存在 {total_failed} 个失败的测试用例", Colors.RED)
        print()
        
        return total_failed == 0
    
    def generate_json_report(self, filename: str = "test_report.json"):
        """生成 JSON 格式报告"""
        report = {
            "start_time": self.start_time.isoformat(),
            "end_time": datetime.now().isoformat(),
            "duration_seconds": (datetime.now() - self.start_time).total_seconds(),
            "test_suites": []
        }
        
        for suite in self.test_suites:
            suite_data = {
                "name": suite.name,
                "total": suite.total,
                "passed": suite.passed,
                "failed": suite.failed,
                "skipped": suite.skipped,
                "pass_rate": suite.pass_rate,
                "tests": [
                    {
                        "name": r.name,
                        "status": r.status.value,
                        "http_code": r.http_code,
                        "response_time": r.response_time,
                        "error_message": r.error_message,
                        "request_url": r.request_url
                    }
                    for r in suite.results
                ]
            }
            report["test_suites"].append(suite_data)
        
        with open(filename, 'w', encoding='utf-8') as f:
            json.dump(report, f, indent=2, ensure_ascii=False)
        
        print(f"JSON 报告已生成: {filename}")
    
    # ==================== 主流程 ====================
    
    def run_all_tests(self):
        """运行所有测试"""
        print(f"\n{Colors.BOLD}Admin API 接口自动化测试{Colors.NC}")
        print(f"API 地址: {self.base_url}")
        print(f"测试账号: {self.username}")
        print("=" * 60)
        
        # 登录
        if not self.login():
            self.print_colored("\n⚠ 登录失败，无法继续测试", Colors.RED)
            return False
        
        # 执行测试套件
        try:
            self.test_system_health()
            self.test_authentication()
            self.test_user_management()
            self.test_rbac()
            self.test_basic_game_config()
            self.test_metadata_definitions()
            self.test_skill_system()
            self.test_effect_system()
            self.test_action_system()
            self.test_error_handling()
        except KeyboardInterrupt:
            self.print_colored("\n\n⚠ 测试被用户中断", Colors.YELLOW)
            return False
        except Exception as e:
            self.print_colored(f"\n\n⚠ 测试执行异常: {str(e)}", Colors.RED)
            import traceback
            traceback.print_exc()
            return False
        
        # 生成报告
        success = self.generate_console_report()
        
        # 生成 JSON 报告
        timestamp = int(time.time())
        json_file = f"test_results_{timestamp}/test_report.json"
        import os
        os.makedirs(os.path.dirname(json_file), exist_ok=True)
        self.generate_json_report(json_file)
        
        return success


def main():
    """主函数"""
    parser = argparse.ArgumentParser(
        description="Admin API 接口自动化测试",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例:
  python3 admin-api-test.py
  python3 admin-api-test.py --url http://localhost:80
  python3 admin-api-test.py --username root --password password
        """
    )
    
    parser.add_argument(
        '--url',
        default='http://localhost:80',
        help='API 基础地址 (默认: http://localhost:80)'
    )
    
    parser.add_argument(
        '--username',
        default='root',
        help='测试账号 (默认: root)'
    )
    
    parser.add_argument(
        '--password',
        default='password',
        help='测试密码 (默认: password)'
    )
    
    args = parser.parse_args()
    
    # 创建测试器并运行
    tester = AdminAPITester(
        base_url=args.url,
        username=args.username,
        password=args.password
    )
    
    success = tester.run_all_tests()
    
    # 根据测试结果返回退出码
    sys.exit(0 if success else 1)


if __name__ == '__main__':
    main()

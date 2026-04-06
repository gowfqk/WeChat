#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
获取企业微信外部联系人ID的脚本
使用方法：python get_external_userids.py
"""

import requests
import json

# 企业微信配置
CORP_ID = "ww0a5b8d4134eff2f9"
CORP_SECRET = "3AoCldfNpZfRLDCxV-m6QLZfQWohx0PA_ei6egSKCx0"

def get_access_token():
    """获取企业微信access_token"""
    url = f"https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid={CORP_ID}&corpsecret={CORP_SECRET}"
    response = requests.get(url)
    result = response.json()

    if result.get('errcode') == 0:
        print("✅ 成功获取access_token")
        return result.get('access_token')
    else:
        print(f"❌ 获取access_token失败: {result}")
        return None

def get_follow_user_list(access_token):
    """获取配置了客户联系功能的成员列表"""
    url = f"https://qyapi.weixin.qq.com/cgi-bin/externalcontact/get_follow_user_list?access_token={access_token}"
    response = requests.post(url)
    result = response.json()

    if result.get('errcode') == 0:
        print(f"✅ 成功获取成员列表，共 {len(result.get('follow_user', []))} 个成员")
        return result.get('follow_user', [])
    else:
        print(f"❌ 获取成员列表失败: {result}")
        return []

def get_contact_list(access_token, userid):
    """获取指定成员的客户列表"""
    url = f"https://qyapi.weixin.qq.com/cgi-bin/externalcontact/list?access_token={access_token}&userid={userid}"
    response = requests.get(url)
    result = response.json()

    if result.get('errcode') == 0:
        external_userid_list = result.get('external_userid', [])
        print(f"✅ 成功获取成员 {userid} 的客户列表，共 {len(external_userid_list)} 个客户")
        return external_userid_list
    else:
        print(f"❌ 获取客户列表失败: {result}")
        return []

def get_contact_detail(access_token, external_userid):
    """获取客户详情"""
    url = f"https://qyapi.weixin.qq.com/cgi-bin/externalcontact/get?access_token={access_token}&external_userid={external_userid}"
    response = requests.get(url)
    result = response.json()

    if result.get('errcode') == 0:
        external_contact = result.get('external_contact', {})
        return external_contact
    else:
        print(f"❌ 获取客户详情失败: {result}")
        return None

def main():
    print("=" * 50)
    print("开始获取企业微信外部联系人ID")
    print("=" * 50)

    # 1. 获取access_token
    print("\n步骤1: 获取access_token...")
    access_token = get_access_token()
    if not access_token:
        print("❌ 无法获取access_token，请检查配置")
        return

    # 2. 获取配置了客户联系功能的成员列表
    print("\n步骤2: 获取配置了客户联系功能的成员列表...")
    follow_users = get_follow_user_list(access_token)
    if not follow_users:
        print("❌ 没有找到配置了客户联系功能的成员")
        return

    print("\n成员列表:")
    for i, user in enumerate(follow_users, 1):
        print(f"  {i}. UserID: {user['userid']}, Name: {user.get('name', 'N/A')}")

    # 3. 获取所有成员的客户列表
    print("\n步骤3: 获取所有成员的客户列表...")
    all_external_userids = []
    contact_details = []

    for user in follow_users:
        userid = user['userid']
        print(f"\n正在获取成员 {userid} 的客户列表...")

        external_userid_list = get_contact_list(access_token, userid)

        if external_userid_list:
            all_external_userids.extend(external_userid_list)

            # 获取每个客户的详细信息
            for ext_userid in external_userid_list:
                print(f"  正在获取客户 {ext_userid} 的详细信息...")
                detail = get_contact_detail(access_token, ext_userid)
                if detail:
                    contact_info = {
                        'external_userid': ext_userid,
                        'name': detail.get('name', 'N/A'),
                        'position': detail.get('position', 'N/A'),
                        'corp_name': detail.get('corp_name', 'N/A'),
                        'type': detail.get('type', 'N/A'),
                        'avatar': detail.get('avatar', 'N/A'),
                        'gender': detail.get('gender', 'N/A'),
                        'unionid': detail.get('unionid', 'N/A'),
                    }
                    contact_details.append(contact_info)
                    print(f"    姓名: {detail.get('name', 'N/A')}, 公司: {detail.get('corp_name', 'N/A')}")

    # 4. 输出结果
    print("\n" + "=" * 50)
    print("获取完成！")
    print("=" * 50)
    print(f"\n总共有 {len(all_external_userids)} 个外部联系人")
    print(f"总共有 {len(follow_users)} 个成员配置了客户联系功能")

    # 保存到文件
    print("\n正在保存结果到文件...")

    # 保存external_userid列表
    with open('external_userids.txt', 'w', encoding='utf-8') as f:
        f.write("外部联系人ID列表:\n")
        f.write("=" * 50 + "\n\n")
        for userid in all_external_userids:
            f.write(f"{userid}\n")

    # 保存详细信息到JSON
    with open('external_contacts_detail.json', 'w', encoding='utf-8') as f:
        json.dump({
            'total_count': len(all_external_userids),
            'total_members': len(follow_users),
            'follow_users': follow_users,
            'external_userids': all_external_userids,
            'contact_details': contact_details
        }, f, ensure_ascii=False, indent=2)

    # 保存详细信息到可读格式
    with open('external_contacts_detail.txt', 'w', encoding='utf-8') as f:
        f.write("外部联系人详细信息\n")
        f.write("=" * 50 + "\n\n")

        for i, contact in enumerate(contact_details, 1):
            f.write(f"{i}. 外部联系人ID: {contact['external_userid']}\n")
            f.write(f"   姓名: {contact['name']}\n")
            f.write(f"   职位: {contact['position']}\n")
            f.write(f"   公司: {contact['corp_name']}\n")
            f.write(f"   类型: {contact['type']}\n")
            f.write(f"   性别: {contact['gender']}\n")
            f.write(f"   UnionID: {contact['unionid']}\n")
            f.write(f"   头像: {contact['avatar']}\n")
            f.write("-" * 50 + "\n")

    print("✅ 结果已保存到以下文件:")
    print("  1. external_userids.txt - 外部联系人ID列表")
    print("  2. external_contacts_detail.json - 详细信息(JSON格式)")
    print("  3. external_contacts_detail.txt - 详细信息(文本格式)")

    # 显示部分结果
    print("\n" + "=" * 50)
    print("部分外部联系人列表:")
    print("=" * 50)
    for i, userid in enumerate(all_external_userids[:10], 1):
        print(f"{i}. {userid}")

    if len(all_external_userids) > 10:
        print(f"... 还有 {len(all_external_userids) - 10} 个")

    print("\n使用示例:")
    print(f'curl -X POST http://localhost:8080/external \\')
    print(f'  -H "Content-Type: application/json" \\')
    print(f'  -d \'{{')
    print(f'    "sendkey": "204800",')
    print(f'    "external_userid": ["{all_external_userids[0] if all_external_userids else "xxx"}"],')
    print(f'    "sender": "{follow_users[0]["userid"] if follow_users else "xxx"}",')
    print(f'    "msgtype": "text",')
    print(f'    "text": {{"content": "测试消息"}}')
    print(f'  }}\''')

if __name__ == "__main__":
    main()

# scripts/release_notes_generator.py

import os
import sys
import subprocess
import json
import re
from openai import OpenAI
from collections import defaultdict

# --- Utility Functions ---

def get_last_tag_and_commits():
    """
    获取最新版本的tag名称，以及从上一个版本tag到当前最新版本tag之间的所有commit。
    标签排序使用 -version:refname，以确保语义版本号正确排序。
    """
    try:
        # 获取所有tag，并按版本号降序排序
        # -version:refname 确保 v1.0.0 在 v0.9.0 之前，v1.0.1 在 v1.0.0 之前
        tags_output = subprocess.check_output("git tag --sort=-version:refname", shell=True).decode('utf-8').strip()
        tags = tags_output.split('\n')

        if not tags or tags[0] == '':
            print("No tags found. Cannot generate release notes.", file=sys.stderr)
            sys.exit(1)

        current_tag = tags[0]
        previous_tag = None
        if len(tags) > 1:
            try:
                current_tag_index = tags.index(current_tag)
                if current_tag_index + 1 < len(tags):
                    previous_tag = tags[current_tag_index + 1]
            except ValueError:
                pass # current_tag not in tags list, this shouldn't happen if tags[0] is current_tag
        
        print(f"Current tag: {current_tag}")
        print(f"Previous tag: {previous_tag if previous_tag else 'None'}")

        commit_range = f"{previous_tag}..{current_tag}" if previous_tag else current_tag
        
        commits_output = subprocess.check_output(
            f'git log {commit_range} --pretty=format:"%H%n%s%n%b%n%an%n%ae%n---COMMIT-END---" --no-merges --grep="^Release" --grep="^Merge branch" --invert-grep',
            shell=True
        ).decode('utf-8').strip()

        commits_raw = commits_output.split('---COMMIT-END---')
        
        commits = []
        for commit_block in commits_raw:
            if not commit_block.strip():
                continue
            lines = commit_block.strip().split('\n')
            if len(lines) >= 4: # Expecting 4 lines for hash, subject, body, author, email
                commit_hash = lines[0]
                subject = lines[1]
                body = "\n".join(lines[2:-2]) if len(lines) > 4 else ""
                author_name = lines[-2]
                author_email = lines[-1]
                
                category = "Other"
                subject_lower = subject.lower()

                if subject_lower.startswith("feat"):
                    category = "Features Added"
                elif subject_lower.startswith("fix") or "bugfix" in subject_lower:
                    category = "Bugs Fixed"
                elif subject_lower.startswith("chore"):
                    category = "Chore"
                elif subject_lower.startswith("docs") or "doc" in subject_lower:
                    category = "Documentation"
                elif subject_lower.startswith("style"):
                    category = "Code Style"
                elif subject_lower.startswith("refactor"):
                    category = "Refactor"
                
                commits.append({
                    "hash": commit_hash,
                    "subject": subject,
                    "body": body,
                    "category": category,
                    "author_name": author_name,
                    "author_email": author_email
                })
        
        return current_tag, commits

    except subprocess.CalledProcessError as e:
        print(f"Error running git command: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"An unexpected error occurred in get_last_tag_and_commits: {e}", file=sys.stderr)
        sys.exit(1)

def get_openai_client():
    """获取OpenAI客户端实例."""
    return OpenAI(
        base_url=os.environ.get("OPENAI_API_BASE_URL"),
        api_key=os.environ.get("OPENAI_API_KEY")
    )

def generate_ai_summary(commits):
    """使用AI生成发布摘要."""
    try:
        client = get_openai_client()
        model = os.environ.get("OPENAI_MODEL", "gpt-4-turbo")

        if not commits:
            return "No significant changes or new features were introduced in this release based on commit history."

        commit_messages_for_ai = []
        for commit in commits:
            commit_messages_for_ai.append(f"Hash: {commit['hash'][:7]}\nSubject: {commit['subject']}\nBody: {commit['body']}")

        prompt = (
            "Based on the following Git commit messages, provide a concise and engaging release summary. "
            "This summary is for a general audience, including non-developers. "
            "Focus on the user-facing impact, new benefits, and improvements. Avoid technical jargon. "
            "Keep it under 200 words. "
            "Output ONLY the summary, with no introductory or concluding remarks or conversational text.\n\n"
            "Commit messages:\n"
            f"{'---\n'.join(commit_messages_for_ai)}"
        )

        print("Sending request for overall AI summary...")
        response = client.chat.completions.create(
            model=model,
            messages=[
                {"role": "system", "content": "You are a helpful assistant. Output strictly adheres to user instructions."},
                {"role": "user", "content": prompt}
            ],
            max_tokens=200,
            temperature=0.7,
        )
        summary = response.choices[0].message.content.strip()
        print("Overall AI summary generated successfully.")
        return summary
    except Exception as e:
        print(f"Error generating AI summary: {e}", file=sys.stderr)
        return "AI generated release summary is not available due to an error."

def generate_ai_formatted_items(category_name, commit_list):
    """
    使用AI将给定分类下的commit列表转换为格式化的发布说明条目。
    AI会处理首字母大写、合并相关commit以及生成可读的描述。
    """
    if not commit_list:
        return ""

    client = get_openai_client()
    model = os.environ.get("OPENAI_MODEL", "gpt-4-turbo")
    repo_url_base = f"https://github.com/{os.environ.get('GITHUB_REPOSITORY')}/commit/"

    commit_data_for_ai = []
    for commit in commit_list:
        commit_data_for_ai.append(f"Hash: {commit['hash']}\nSubject: {commit['subject']}\nBody: {commit['body']}\n")
    
    prompt = (
        f"Convert the following raw Git commit messages for the '{category_name}' section "
        "into a user-friendly, concise, and well-formatted Markdown list for release notes. "
        "Each item should describe a change from a user's perspective, avoiding technical implementation details. "
        "Output ONLY the Markdown list, with no other text, introduction, or conclusion or conversational text.\n\n"
        "Rules:\n"
        "1. Combine related commits into a single, comprehensive list item if they describe parts of the same feature or fix.\n"
        "2. For each list item, summarize the change in a clear, concise, and customer-facing manner. Start the description with a capital letter.\n"
        "3. Include the first 6 characters of ALL relevant commit SHAs for that item, formatted as `[`hash`]` and linked. "
        "   If multiple commits contribute to one item, list all their SHAs separated by `, ` (comma and space). "
        f"   Example: `[`first6`](LINK_TO_COMMIT_1), [`second6`](LINK_TO_COMMIT_2)` where LINK_TO_COMMIT is `{repo_url_base}<full_hash>`.\n"
        "4. Each item starts with `-`.\n"
        "5. If a commit message is unclear or internal, infer its user-facing impact or omit if irrelevant.\n\n"
        "Raw commit messages:\n"
        f"{'---\n'.join(commit_data_for_ai)}"
    )

    print(f"Sending request for AI formatted items for category: {category_name}...")
    try:
        response = client.chat.completions.create(
            model=model,
            messages=[
                {"role": "system", "content": "You are a helpful assistant that summarizes Git commits into release notes. Output strictly adheres to user instructions."},
                {"role": "user", "content": prompt}
            ],
            max_tokens=1024,
            temperature=0.3,
        )
        formatted_items = response.choices[0].message.content.strip()

        processed_items = []
        for line in formatted_items.split('\n'):
            line = line.strip()
            if not line: # Skip empty lines
                continue
            
            hashes_in_line = re.findall(r'`\[([a-fA-F0-9,\s]+?)\]`', line)
            
            formatted_hash_links = []
            if hashes_in_line:
                all_hashes_found = []
                for h_group in hashes_in_line:
                    all_hashes_found.extend([h.strip() for h in h_group.split(',') if h.strip()])
                
                for h_short in all_hashes_found:
                    formatted_hash_links.append(f"[`{h_short}`]({repo_url_base}{h_short})")
                
                if line.startswith('- '):
                    match = re.match(r'^-?\s*(`\[[a-fA-F0-9,\s]+?`\](\([^\)]+\))?(?:,\s*`\[[a-fA-F0-9,\s]+?`\](\([^\)]+\))?)*):\s*(.*)', line)
                    if match:
                        description_part = match.group(5).strip()
                        prefix = ", ".join(formatted_hash_links)
                        processed_items.append(f"- {prefix}: {description_part}")
                    else:
                        temp_line = line
                        for h_short in all_hashes_found:
                            temp_line = re.sub(r'`?\[' + re.escape(h_short) + r'\]`?', f"[`{h_short}`]({repo_url_base}{h_short})", temp_line)
                        processed_items.append(temp_line)
                else:
                    processed_items.append(line)
            else:
                processed_items.append(line)
        
        print(f"AI formatted items generated for {category_name}.")
        return "\n".join(processed_items) + "\n" if processed_items else ""

    except Exception as e:
        print(f"Error generating AI formatted items for {category_name}: {e}", file=sys.stderr)
        return "- AI failed to generate details for this section.\n"

def get_github_username_from_email(email):
    """
    尝试从邮件地址中提取GitHub用户名。
    支持 format: ID+username@users.noreply.github.com
    """
    match = re.search(r'\+(\w+)@users\.noreply\.github\.com', email)
    if match:
        return match.group(1)
    return None

def generate_contributors_section(commits, max_columns=7):
    """
    生成贡献者部分，使用 Markdown 表格，精确到只有实际贡献者的列数。
    尝试创建 3 行的表格：头像，用户名，贡献数。
    """
    contributor_counts = defaultdict(int)
    contributor_info = {} # Stores {author_email: {name, github_username, avatar_url}}

    for commit in commits:
        author_email = commit['author_email']
        author_name = commit['author_name']
        contributor_counts[author_email] += 1

        if author_email not in contributor_info:
            github_username = get_github_username_from_email(author_email)
            avatar_url = "https://github.com/github.png?size=40" # Default generic avatar
            
            if github_username:
                avatar_url = f"https://github.com/{github_username}.png?size=40"
            
            contributor_info[author_email] = {
                "name": author_name,
                "github_username": github_username, # Keep original if not from GitHub email
                "avatar_url": avatar_url
            }

    # Sort contributors by commit count in descending order
    sorted_contributors = sorted(contributor_counts.items(), key=lambda item: item[1], reverse=True)

    if not sorted_contributors:
        return ""

    contributors_markdown = "## Contributors\n\nSpecial thanks to:\n\n"

    # Split contributors into chunks of max_columns
    for i in range(0, len(sorted_contributors), max_columns):
        chunk = sorted_contributors[i:i + max_columns]
        
        # Initialize rows for the current chunk
        image_row_cells = []
        separator_row_cells = []
        name_row_cells = []
        count_row_cells = []

        for email, count in chunk:
            info = contributor_info[email]
            display_name = info["github_username"] if info["github_username"] else info["name"]
            
            image_row_cells.append(f"![Avatar]({info['avatar_url']}) ") # Space before | for better rendering
            separator_row_cells.append(" :----------: ") # Centered alignment
            name_row_cells.append(f" **{display_name}** ")
            count_row_cells.append(f" {count} commit{'s' if count > 1 else ''} ")

        # Join cells with "|" and add leading/trailing "|" for the table
        contributors_markdown += "|" + "|".join(image_row_cells) + "|\n"
        contributors_markdown += "|" + "|".join(separator_row_cells) + "|\n"
        contributors_markdown += "|" + "|".join(name_row_cells) + "|\n"
        contributors_markdown += "|" + "|".join(count_row_cells) + "|\n\n" # Two newlines to separate tables if multiple chunks

    return contributors_markdown

def format_release_notes(tag_name, ai_summary, all_commits):
    """格式化发布说明."""
    release_notes = f"## Release {tag_name}\n\n" # 标题后两个空行
    release_notes += f"{ai_summary}\n\n" # 摘要后两个空行

    # Add Contributors section at the top, after the main summary
    contributors_section = generate_contributors_section(all_commits)
    if contributors_section:
        release_notes += contributors_section

    categorized_commits = {
        "Features Added": [],
        "Bugs Fixed": [],
        "Chore": [],
        "Documentation": [],
        "Code Style": [],
        "Refactor": [],
        "Other": []
    }

    for commit in all_commits:
        category = commit.get("category", "Other")
        categorized_commits[category].append(commit)

    for category, commit_list in categorized_commits.items():
        if commit_list:
            release_notes += f"### {category}\n\n" # 每个分类标题后两个空行
            ai_items = generate_ai_formatted_items(category, commit_list)
            release_notes += ai_items
            
    # Add the footer
    release_notes += "---\n\nMade with ♥️ by Tethys Plex & Veloera.\n"
    
    return release_notes

def main():
    tag_name, commits = get_last_tag_and_commits()

    if not tag_name:
        print("No tag found to generate release notes for.", file=sys.stderr)
        sys.exit(1)

    ai_summary = generate_ai_summary(commits)

    release_notes_content = format_release_notes(tag_name, ai_summary, commits)

    output_dir = "docs/release-notes"
    os.makedirs(output_dir, exist_ok=True)

    output_filename = os.path.join(output_dir, f"{tag_name}.md")
    with open(output_filename, "w", encoding="utf-8") as f:
        f.write(release_notes_content)

    print(f"Release notes for {tag_name} saved to {output_filename}")
    
    # Set output for GitHub Actions using GITHUB_OUTPUT environment file
    if 'GITHUB_OUTPUT' in os.environ:
        with open(os.environ['GITHUB_OUTPUT'], 'a') as f:
            f.write(f"release_notes_file={output_filename}\n")
            f.write(f"tag_name={tag_name}\n")
    else:
        print("GITHUB_OUTPUT environment variable not found. Skipping setting workflow outputs (expected when running locally).")

if __name__ == "__main__":
    main()
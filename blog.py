from flask import Blueprint, render_template, abort, url_for
from database import db, BlogPost
import json

blog_bp = Blueprint('blog', __name__)

@blog_bp.route('/blog')
def blog_list():
    try:
        posts = BlogPost.query.order_by(BlogPost.created_at.desc()).all()
    except Exception as e:
        db.session.rollback()
        posts = []
        print(f"Error fetching blog posts: {e}")
    return render_template('blog.html', posts=posts)

@blog_bp.route('/blog/<slug>')
def blog_detail(slug):
    post = BlogPost.query.filter_by(slug=slug).first_or_404()
    
    # Increment views
    try:
        post.views += 1
        db.session.commit()
    except:
        db.session.rollback()

    # Prepare JSON-LD Schema.org data for this specific blog post
    schema_data = {
        "@context": "https://schema.org",
        "@type": "BlogPosting",
        "headline": post.title,
        "description": post.summary,
        "author": {
            "@type": "Person",
            "name": "Jonathan Roets"
        },
        "datePublished": post.created_at.isoformat(),
        "url": url_for('blog.blog_detail', slug=post.slug, _external=True)
    }

    return render_template(
        'blog_detail.html',
        post=post,
        seo_title=f"{post.title} | JO4 Dev Blog",
        seo_description=post.summary,
        json_ld_schema=json.dumps(schema_data, indent=2)
    )
